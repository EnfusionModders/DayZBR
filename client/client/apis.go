package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/dbhelpers"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/discord-bot/discord"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/steam-query/steamquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Seq struct {
		Endpoint string `yaml:"endpoint"`
		Apikey   string `yaml:"apikey"`
		Appname  string `yaml:"appname"`
		Loglevel uint32 `yaml:"loglevel"`
		Seqlevel uint32 `yaml:"seqlevel"`
	} `yaml:"seq"`
	Discord struct {
		Token         string `yaml:"token"`
		DebugChannel  string `yaml:"debug_channel"`
		StatusChannel string `yaml:"status_channel"`
	} `yaml:"discord"`
	Mongodb struct {
		ConnectUri string `yaml:"connect_uri"`
	} `yaml:"mongodb"`
}
type Client struct {
	Log     *logrus.Entry
	Cfg     *Config
	Metrics *ClientMetrics
	Bot     *discord.DiscordBot
	Db      *database.BattleRoyaleDB
}
type ClientMetrics struct {
	clients_started      prometheus.Counter
	matchmakes_started   prometheus.Counter
	matchmakes_completed prometheus.Counter
}

func NewClient(log *logrus.Entry, cfg *Config) *Client {
	bot, err := discord.NewDiscordBot(cfg.Discord.Token, log)
	if err != nil {
		log.Errorln("Failed to initialize discord bot", err)
		return nil
	}
	db, err := database.ConnectTo(cfg.Mongodb.ConnectUri)
	if err != nil {
		log.Errorln("Failed to initialize database connection", err)
		return nil
	}
	return &Client{
		Log: log,
		Cfg: cfg,
		Bot: bot,
		Db:  db,
		Metrics: &ClientMetrics{
			clients_started: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "clientservice",
				Name:      "clientsstarted",
				Help:      "Total calls to start.",
			}),
			matchmakes_started: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "clientservice",
				Name:      "matchmakesstarted",
				Help:      "Total calls to matchmake.",
			}),
			matchmakes_completed: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "clientservice",
				Name:      "matchmakesfinished",
				Help:      "Total completed matchmakes.",
			}),
		},
	}
}

func (c *Client) Ping(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "pong")
}
func (c *Client) Start(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")

	req, err := ParseStartRequest(r.Body)
	if err != nil {
		c.Log.Errorln("Failed to parse Start request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}
	if len(req.SteamId) != 17 {
		if len(req.SteamId) > 24 {
			c.Log.Errorln("invalid steamid. too long")
		} else {
			c.Log.Errorln("invalid steamid", req.SteamId)
		}
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	//FindPlayer by steam id
	plr, err := c.Db.GetPlayerBySteamID(req.SteamId)
	if err != nil {

		if err != mongo.ErrNoDocuments {
			c.Log.Errorln("Failed to find player in db", err)
			json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
			return
		}
		//player not found? insert
		plr, err = c.Db.InsertPlayer(&database.BRPlayer{
			Name:      req.Name,
			SteamId:   req.SteamId,
			Ips:       []string{ip},
			Purchases: make([]string, 0),
		})
		if err != nil {
			c.Log.Errorln("Failed to insert player into db", err)
			json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
			return
		}
	} else {
		//update player ip if necessary
		found_ip := false
		for _, dbip := range plr.Ips {
			if dbip == ip {
				found_ip = true
			}
		}
		if !found_ip || strings.Trim(strings.ToLower(plr.Name), " \t\n\rr") != strings.Trim(strings.ToLower(req.Name), " \t\n\r") {
			plr.Ips = append(plr.Ips, ip)
			plr.Name = req.Name
			plr, err = c.Db.UpdatePlayer(plr)
			if err != nil {
				c.Log.Errorln("Failed to update player ip", err)
				json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
				return
			}
		}
	}

	regions, err := dbhelpers.GetAllRegions(c.Db)
	if err != nil {
		c.Log.Errorln("Failed to get all regions", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	c.Log.WithField("num_regions", len(regions)).Debugln("Found {num_regions} region(s).")

	c.Metrics.clients_started.Inc()
	json.NewEncoder(w).Encode(buildSuccessResponse(&StartResponse{
		Player: *plr,
		Region: struct {
			Regions []string "json:\"regions\""
		}{
			Regions: regions,
		},
	}))
}
func (c *Client) Matchmake(w http.ResponseWriter, r *http.Request) {
	req, err := ParseMatchMakeRequest(r.Body)
	if err != nil {
		c.Log.Errorln("Failed to parse Matchmake request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	//This is unused, but could be required in the future to help better place players when matchmaking
	/*
		plr, err := c.Db.GetPlayerByID(req.PlayerId)
		if err != nil {
			c.Log.Errorln("Failed to get player by id", err)
			json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
			return
		}
	*/

	regions := []string{"any"}
	if req.Region != "" {
		regions = append(regions, req.Region)
	}
	svrs, err := c.Db.GetServersFiltered(bson.M{
		"locked":  0,
		"version": req.Version,
		"region": bson.M{
			"$in": regions,
		},
	})
	if err != nil {
		c.Log.Errorln("Failed to get servers", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}
	//no servers found, send back false for wait and invalid server object
	if len(svrs) == 0 {
		c.Log.WithFields(logrus.Fields{
			"version": req.Version,
			"region":  req.Region,
		}).Warnln("No servers found for version {version} in region {region}")
		json.NewEncoder(w).Encode(buildSuccessResponse(&MatchmakeResponse{
			Wait:   0,
			Server: database.BRServer{},
		}))
		return
	}

	var results []*MatchMakeQuery

	c.Metrics.matchmakes_started.Inc()

	var wg sync.WaitGroup
	for _, svr := range svrs {
		if svr == nil {
			continue
		}
		//we are using a goroutine here to calculate optimal servers. This will let us matchmake blazingly fast.
		wg.Add(1)
		go func(server *database.BRServer) {
			defer wg.Done()
			conn := server.Connection
			parts := strings.Split(conn, ":")
			if len(parts) != 2 {
				c.Log.Warnln("Invalid server connection details", conn)
				return
			}

			query, err := steamquery.QueryDayZ(parts[0], server.QueryPort, 5*time.Second)
			if err != nil {
				c.Log.WithFields(logrus.Fields{
					"ip":        parts[0],
					"queryport": server.QueryPort,
				}).Warnln("Failed to query server {ip}:{queryport}")
			}
			results = append(results, &MatchMakeQuery{
				Query:  query,
				Server: server,
			})
		}(svr)
	}
	wg.Wait()

	var optimal_server *MatchMakeQuery
	optimal_percent_full := float32(0)
	for _, res := range results {
		if optimal_server == nil {
			optimal_server = res
		} else {
			percent_full := float32(res.Query.Players) / float32(res.Query.Players)
			//TODO: add more filters for determining ideal matchmaked server
			if percent_full > optimal_percent_full {
				optimal_percent_full = percent_full
				optimal_server = res
			}
		}
	}
	if optimal_server == nil {
		//no optimal server (must all be in use) -- send wait signal
		json.NewEncoder(w).Encode(buildSuccessResponse(&MatchmakeResponse{
			Wait:   1,
			Server: database.BRServer{},
		}))
	} else {
		json.NewEncoder(w).Encode(buildSuccessResponse(&MatchmakeResponse{
			Wait:   0,
			Server: *optimal_server.Server,
		}))
	}

	c.Metrics.matchmakes_completed.Inc()
}
func (c *Client) GetPlayer(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetPlayerRequest(r.Body)
	if err != nil {
		c.Log.Errorln("Failed to parse GetPlayer request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	var player database.BRPlayer
	if len(req.PlayerId) == 24 {
		//db id style
		plr, err := c.Db.GetPlayerByID(req.PlayerId)
		if err != nil {
			c.Log.Warnln("Player id not found", err)
			json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
			return
		}
		player = *plr
	} else if len(req.PlayerId) == 17 {
		//steam id style
		plr, err := c.Db.GetPlayerBySteamID(req.PlayerId)
		if err != nil {
			c.Log.Warnln("Player steam id not found", err)
			json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
			return
		}
		player = *plr
	} else {
		c.Log.Errorln("Invalid player id")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(player))
}
func (c *Client) GetMatch(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetMatchRequest(r.Body)
	if err != nil {
		c.Log.Errorln("Failed to parse GetPlayer request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	match, err := c.Db.GetRawMatchByID(req.MatchId)
	if err != nil {
		c.Log.Errorln("Match id not found", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(*match))
}
func (c *Client) GetServer(w http.ResponseWriter, r *http.Request) {
	req, err := ParseGetServerRequest(r.Body)
	if err != nil {
		c.Log.Errorln("Failed to parse GetPlayer request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	svr, err := c.Db.GetServerByID(req.ServerId)
	if err != nil {
		c.Log.Errorln("Server id not found", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(*svr))
}
