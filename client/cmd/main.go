package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/didip/tollbooth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/argparse/argparse"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/client/client"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/configyml/configyml"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/logger/seq"
)

func main() {
	configpath := argparse.GetArg("config", "config.yml")

	var cfg client.Config
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

	log.Infoln("Starting DayZBR Client Service")

	cli := client.NewClient(log, &cfg)

	lmt := tollbooth.NewLimiter(50, nil) //50 requests per second
	lmt.SetMessage("Too many requests.")

	m := pat.New()

	m.Get("/client/ping", tollbooth.LimitFuncHandler(lmt, cli.Ping))

	m.Post("/client/start", tollbooth.LimitFuncHandler(lmt, cli.Start))
	m.Post("/client/matchmake", tollbooth.LimitFuncHandler(lmt, cli.Matchmake))
	m.Post("/client/player", tollbooth.LimitFuncHandler(lmt, cli.GetPlayer))
	m.Post("/client/match", tollbooth.LimitFuncHandler(lmt, cli.GetMatch))
	m.Post("/client/server", tollbooth.LimitFuncHandler(lmt, cli.GetServer))

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
	io.WriteString(w, "DayZBR Client Service.")
}
func Health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Healthy!\n")
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Endpoint not found!")
}
