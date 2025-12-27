package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/discord-bot/discord"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/steam-query/steamquery"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Server struct {
		Port       string `yaml:"port"`
		PrivateKey string `yaml:"private_key"`
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

type Server struct {
	Log     *logrus.Entry
	Cfg     *Config
	Metrics *ServerMetrics
	Bot     *discord.DiscordBot
	Db      *database.BattleRoyaleDB
}
type ServerMetrics struct {
	total_started  prometheus.Counter
	total_finished prometheus.Counter
}

func NewServer(log *logrus.Entry, cfg *Config) *Server {
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
	return &Server{
		Log: log,
		Cfg: cfg,
		Bot: bot,
		Db:  db,
		Metrics: &ServerMetrics{
			total_started: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "serverservice",
				Name:      "totalstarted",
				Help:      "Total calls to onstart.",
			}),
			total_finished: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "serverservice",
				Name:      "totalfinished",
				Help:      "Total calls to onfinish.",
			}),
		},
	}
}

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	if !s.validateKey(w, r) {
		//invalid private key!
		io.WriteString(w, "invalid request")
		return
	}
	io.WriteString(w, "Pong!")
}
func (s *Server) BotTest(w http.ResponseWriter, r *http.Request) {
	if !s.validateKey(w, r) {
		//invalid private key!
		io.WriteString(w, "invalid request")
	}
	io.WriteString(w, "testing discord bot...")

	//TODO: discord bot integration!
}
func (s *Server) OnStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !s.validateKey(w, r) {
		json.NewEncoder(w).Encode(buildErrorResponse("invalid access key"))
		s.Log.Debugln("Invalid private key in request", r.URL.RawQuery)
		return
	}
	//called on server startup
	req, err := ParseOnStartRequest(r.Body)
	if err != nil {
		s.Log.Errorln("Failed to parse OnStart request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	if req.ServerIP == "localhost" || req.ServerIP == "127.0.0.1" {
		req.ServerIP = r.Header.Get("X-Forwarded-For")
	}

	// SteamQuery synchronously
	query, err := steamquery.QueryDayZ(req.ServerIP, req.QueryPort, time.Second*5)
	if err != nil {
		s.Log.Errorln("Failed to query dayz server", err)
		json.NewEncoder(w).Encode(buildErrorResponse("failed to reach server"))
		return
	}

	s.Bot.MessageChannel(s.Cfg.Discord.StatusChannel, query.ServerName+" is now online! Map: "+query.Map)

	//find server in the db
	svr, err := s.Db.GetServerByConnection(req.ServerIP, req.QueryPort)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			s.Log.Errorln("Failed to query database for server", err)
			json.NewEncoder(w).Encode(buildErrorResponse("internal error"))
			return
		}

		//no server with details exists -- insert
		svr := &database.BRServer{
			Name:       query.ServerName,
			Connection: req.ServerIP + ":" + strconv.Itoa(query.Port),
			QueryPort:  req.QueryPort,
			Version:    req.ServerVersion,
			Matches:    make([]primitive.ObjectID, 0),
			Region:     "any",
			Locked:     1,
		}
		svr, err = s.Db.InsertServer(svr)
		if err != nil {
			s.Log.Errorln("Failed to update server in database", err)
			json.NewEncoder(w).Encode(buildErrorResponse("internal error"))
			return
		}

		//return svr as json
		s.Metrics.total_started.Inc()
		json.NewEncoder(w).Encode(buildSuccessResponse(*svr))
	} else {
		//update server details
		svr.Connection = req.ServerIP + ":" + strconv.Itoa(query.Port)
		svr.LastStarted = time.Now().Unix()
		svr.Name = query.ServerName
		svr.Version = req.ServerVersion
		svr.QueryPort = req.QueryPort

		// update server
		svr, err = s.Db.UpdateServer(svr)
		if err != nil {
			s.Log.Errorln("Failed to update server in database", err)
			json.NewEncoder(w).Encode(buildErrorResponse("internal error"))
			return
		}

		// return svr as json
		s.Metrics.total_started.Inc()
		json.NewEncoder(w).Encode(buildSuccessResponse(*svr)) //need to pass the full object, not a pointer
	}
}
func (s *Server) SetLock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !s.validateKey(w, r) {
		json.NewEncoder(w).Encode(buildErrorResponse("invalid access key"))
		s.Log.Debugln("Invalid private key in request", r.URL.RawQuery)
		return
	}
	//called on server startup
	req, err := ParseSetLockRequest(r.Body)
	if err != nil {
		s.Log.Errorln("Failed to parse SetLock request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	if req.ServerIP == "localhost" || req.ServerIP == "127.0.0.1" {
		req.ServerIP = r.Header.Get("X-Forwarded-For")
	}

	s.Log.Debugln("Getting server by connection", req.ServerIP, "with queryport", req.QueryPort)
	svr, err := s.Db.GetServerByConnection(req.ServerIP, req.QueryPort)
	if err != nil {
		s.Log.Errorln("Failed to find server in db", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	svr.Locked = req.Lock
	svr, err = s.Db.UpdateServer(svr)
	if err != nil {
		s.Log.Errorln("Failed to update server lock", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(nil))

}
func (s *Server) OnFinish(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !s.validateKey(w, r) {
		json.NewEncoder(w).Encode(buildErrorResponse("invalid access key"))
		s.Log.Debugln("Invalid private key in request", r.URL.RawQuery)
		return
	}
	//called on server startup
	req, err := ParseOnFinishRequest(r.Body)
	if err != nil {
		s.Log.Errorln("Failed to parse OnFinish request", err)
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	if req.ServerIP == "localhost" || req.ServerIP == "127.0.0.1" {
		req.ServerIP = r.Header.Get("X-Forwarded-For")
	}

	// SteamQuery synchronously
	query, err := steamquery.QueryDayZ(req.ServerIP, req.QueryPort, time.Second*5)
	if err != nil {
		s.Log.Errorln("Failed to query dayz server", err)
		json.NewEncoder(w).Encode(buildErrorResponse("failed to reach server"))
		return
	}

	s.Bot.MessageChannel(s.Cfg.Discord.StatusChannel, query.ServerName+" has finished it's match! "+req.Winner+" has won!")
	s.Metrics.total_finished.Inc()
	json.NewEncoder(w).Encode(buildSuccessResponse(nil))
}

func (s *Server) validateKey(w http.ResponseWriter, r *http.Request) bool {
	private_key := r.URL.Query().Get(":privkey")

	if s.Cfg.Server.PrivateKey != private_key {
		return false
	}
	return true
}
