package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/didip/tollbooth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/argparse/argparse"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/configyml/configyml"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard/postprocessing"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/logger/seq"
)

func main() {
	configpath := argparse.GetArg("config", "config.yml")
	var cfg leaderboard.Config
	err := configyml.GetConfig(configpath, &cfg)
	if err != nil {
		fmt.Println("failed to read config", err)
		return
	}

	log := seq.NewLogger(&seq.SeqConfig{
		Endpoint: cfg.Seq.Endpoint,
		Apikey:   cfg.Seq.Apikey,
		Appname:  cfg.Seq.Appname,
		Loglevel: cfg.Seq.Loglevel,
		Seqlevel: cfg.Seq.Seqlevel,
	})

	log.Infoln("Starting DayZBR Leaderboard Service")

	ldr := leaderboard.NewLeaderboard(log, &cfg)
	if ldr == nil {
		return
	}

	go postprocessing.ProcessNewMatchRoutine(ldr) //start match postprocess goroutine

	//rate limiter
	lmt := tollbooth.NewLimiter(50, nil) //50 requests per second
	lmt.SetMessage("Too many requests.")

	//router
	m := pat.New()
	m.Get("/data/:privkey/ping", tollbooth.LimitFuncHandler(lmt, ldr.Ping))

	m.Post("/data/:privkey/matchsubmit", tollbooth.LimitFuncHandler(lmt, ldr.SubmitMatch))
	m.Post("/data/steaminfo", tollbooth.LimitFuncHandler(lmt, ldr.GetPlayerSteamInfo))
	m.Post("/data/player", tollbooth.LimitFuncHandler(lmt, ldr.GetPlayerData))
	m.Post("/data/match", tollbooth.LimitFuncHandler(lmt, ldr.GetMatchData))
	m.Post("/data/global", tollbooth.LimitFuncHandler(lmt, ldr.GetGlobalData))
	m.Post("/data/rank", tollbooth.LimitFuncHandler(lmt, ldr.GetGlobalRank))

	m.Get("/metrics", promhttp.Handler())
	m.Post("/metrics", promhttp.Handler())

	m.Get("/", http.HandlerFunc(Root))
	m.Get("/health", http.HandlerFunc(Root))
	m.NotFound = http.HandlerFunc(NotFound)

	http.Handle("/", m)
	err = http.ListenAndServe(":"+cfg.Server.Port, nil)
	if err != nil {
		log.WithError(err).Errorln("Failed to listen and serve. {error}")
	}
}

func Root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "DayZBR Leaderboard Data Service.")
}
func Health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Healthy!\n")
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Endpoint not found!")
}
