package leaderboard

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/discord-bot/discord"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard/steam"
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
	Steam struct {
		ApiKey string `yaml:"apikey"`
	} `yaml:"steam"`
}

type Leaderboard struct {
	Log     *logrus.Entry
	Cfg     *Config
	Metrics *LeaderboardMetrics
	Bot     *discord.DiscordBot
	Db      *database.BattleRoyaleDB
	Chn     chan *database.BRRawMatch
}
type LeaderboardMetrics struct {
	matches_submitted prometheus.Counter
}

func NewLeaderboard(log *logrus.Entry, cfg *Config) *Leaderboard {

	pp_chan := make(chan *database.BRRawMatch, 100)

	bot, err := discord.NewDiscordBot(cfg.Discord.Token, log)
	if err != nil {
		log.WithError(err).Errorln("Failed to initialize discord bot. {error}")
		return nil
	}
	db, err := database.ConnectTo(cfg.Mongodb.ConnectUri)
	if err != nil {
		log.WithError(err).Errorln("Failed to initialize database connection. {error}")
		return nil
	}
	return &Leaderboard{
		Log: log,
		Cfg: cfg,
		Bot: bot,
		Db:  db,
		Chn: pp_chan,
		Metrics: &LeaderboardMetrics{
			matches_submitted: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "dayzbr",
				Subsystem: "leaderboard",
				Name:      "submissions",
				Help:      "Total matches submitted.",
			}),
		},
	}
}

func (s *Leaderboard) Ping(w http.ResponseWriter, r *http.Request) {
	if !s.validateKey(w, r) {
		//invalid private key!
		io.WriteString(w, "invalid request")
		return
	}
	io.WriteString(w, "Pong!")
}
func (s *Leaderboard) validateKey(w http.ResponseWriter, r *http.Request) bool {
	private_key := r.URL.Query().Get(":privkey")

	if s.Cfg.Server.PrivateKey != private_key {
		return false
	}
	return true
}

//get leaderboards object for player (steamid input)
func (s *Leaderboard) GetPlayerData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	req, err := ParsePlayerDataRequest(r.Body)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse GetPlayerData request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	steamid := req.SteamID
	//TODO: get player leaderboard info for steamid

	player, err := s.Db.GetLeaderboardPlayerBySteamID(steamid)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to get player by steam id. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}
	json.NewEncoder(w).Encode(buildSuccessResponse(*player))
}
func (s *Leaderboard) GetGlobalRank(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	req, err := ParseRankRequest(r.Body)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse GetGlobalRank request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	rating := req.Rating

	rank, err := s.Db.GetLeaderboardRank(rating)
	if err != nil {
		s.Log.WithError(err).WithField("rating", rating).Errorln("Failed to calculate rank from rating. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(RankResponse{
		Rank: int(rank),
	}))
}

//get player steam name, avatar url, and profile link (steamid input)
func (s *Leaderboard) GetPlayerSteamInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	req, err := ParseSteamInfoRequest(r.Body)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse SteamInfo request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	info, err := steam.GetPlayerInfo(req.SteamID, s.Cfg.Steam.ApiKey)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse SubmitMatch request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid steam id"))
		return
	}

	name := info.PersonaName
	pic := info.LargeAvatarURL
	url := info.ProfileURL
	//thse are returned

	res := SteamInfoResponse{
		ProfileName: name,
		ProfilePic:  pic,
		ProfileUrl:  url,
	}
	json.NewEncoder(w).Encode(buildSuccessResponse(res))
	return
}

//get leaderboards obnject for match (match object id input)
func (s *Leaderboard) GetMatchData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	req, err := ParseMatchDataRequest(r.Body)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse GetMatchData request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	//this is the match id (hex) of
	hex := req.MatchID
	is_ldb_match := req.IsLeaderboardMatch

	var match *database.LeaderboardMatch
	if is_ldb_match {
		match, err = s.Db.GetLeaderboardMatchByID(hex)
	} else {
		match, err = s.Db.GetLeaderboardMatchByRawID(hex)
	}
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to get match by id. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	json.NewEncoder(w).Encode(buildSuccessResponse(*match))
}

//get global leaderboards data object
func (s *Leaderboard) GetGlobalData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	global, err := s.Db.GetLeaderboardGlobaldata()
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to get global leaderboard data. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("failed request"))
		return
	}

	top_players, err := s.Db.GetTopLeaderboardPlayersByRating(10)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to get top ranked players. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("failed request"))
		return
	}

	res := GlobalDataResponse{
		Data:    *global,
		Players: top_players,
	}
	json.NewEncoder(w).Encode(buildSuccessResponse(res))
}

func (s *Leaderboard) SubmitMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if !s.validateKey(w, r) {
		json.NewEncoder(w).Encode(buildErrorResponse("invalid access key"))
		s.Log.Debugln("Invalid private key in request", r.URL.RawQuery)
		return
	}
	//called on server startup

	req, err := ParseSubmitMatchRequest(r.Body)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to parse SubmitMatch request. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	svr, err := s.Db.GetServerByID(req.ServerID)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to find server by id. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}
	//update matchdata, provide server id
	if len(req.MatchData.Results) < 1 {
		s.Log.WithField("len", len(req.MatchData.Results)).Errorln("No deaths found in matchdata.")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}
	req.MatchData.ServerId = svr.Id
	match, err := s.Db.InsertRawMatch(&req.MatchData)
	if err != nil {
		s.Log.WithError(err).Errorln("Failed to insert match data. {error}")
		json.NewEncoder(w).Encode(buildErrorResponse("invalid request"))
		return
	}

	s.Bot.MessageChannel(s.Cfg.Discord.DebugChannel, "New match data collected! https://dayzbr.dev/client/match/"+match.Id.Hex())
	s.Bot.MessageChannel(s.Cfg.Discord.StatusChannel, "Match finished! https://dayzbr.dev/leaderboard/match/"+match.Id.Hex())

	//--- push match into postprocessing queue
	s.Chn <- match

	s.Metrics.matches_submitted.Inc()
	json.NewEncoder(w).Encode(buildSuccessResponse(nil))
}
