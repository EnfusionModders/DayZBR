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
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/logger/seq"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/server/server"
)

func main() {
	configpath := argparse.GetArg("config", "config.yml")

	var cfg server.Config
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

	log.Infoln("Starting DayZBR Server Service")

	//server object
	svr := server.NewServer(log, &cfg)
	if svr == nil {
		return //errors would already have been logged
	}

	//rate limiter
	lmt := tollbooth.NewLimiter(50, nil) //50 requests per second
	lmt.SetMessage("Too many requests.")

	//router
	m := pat.New()
	m.Get("/server/:privkey/ping", tollbooth.LimitFuncHandler(lmt, svr.Ping))
	m.Get("/server/:privkey/bot/test", tollbooth.LimitFuncHandler(lmt, svr.BotTest))

	m.Post("/server/:privkey/onstart", tollbooth.LimitFuncHandler(lmt, svr.OnStart))
	m.Post("/server/:privkey/setlock", tollbooth.LimitFuncHandler(lmt, svr.SetLock))
	m.Post("/server/:privkey/onfinish", tollbooth.LimitFuncHandler(lmt, svr.OnFinish))

	m.Get("/metrics", promhttp.Handler())
	m.Post("/metrics", promhttp.Handler())

	m.Get("/", http.HandlerFunc(Root))
	m.Get("/health", http.HandlerFunc(Root))
	m.NotFound = http.HandlerFunc(NotFound)

	http.Handle("/", m)
	err = http.ListenAndServe(":"+cfg.Server.Port, nil)
	if err != nil {
		log.Errorln("Failed to listen and serve", err)
	}
}

func Root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "DayZBR Server Service.")
}
func Health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Healthy!\n")
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Endpoint not found!")
}
