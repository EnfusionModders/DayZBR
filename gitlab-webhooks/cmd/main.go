package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bmizerany/pat"
	"github.com/didip/tollbooth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/argparse/argparse"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/configyml/configyml"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/discord-bot/discord"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/logger/seq"
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
	Gitlab struct {
		Token string `yaml:"secret_token"`
	} `yaml:"gitlab"`
}
type GitlabWebhook struct {
	total_requests prometheus.Counter
	cfg            *Config
	log            *logrus.Entry
	bot            *discord.DiscordBot
}

func main() {
	configpath := argparse.GetArg("config", "config.yml")

	var cfg Config
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

	log.Infoln("Starting DayZBR Gitlab Webhook Service")

	bot, err := discord.NewDiscordBot(cfg.Discord.Token, log)
	if err != nil {
		if bot != nil {
			defer bot.Shutdown() //if bot is open, we need to shut it down
		}
		log.Fatalln("Failed to initialize discord bot", err)
		return
	}

	//discord bot message debug channel
	bot.MessageChannel(cfg.Discord.DebugChannel, "DayZBR Webhooks Service Started!")

	//init our webhook handler object
	webhook := &GitlabWebhook{
		//init prometheus metrics
		total_requests: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "dayzbr",
			Subsystem: "gitlabwebhooks",
			Name:      "requests",
			Help:      "total valid webhook requests",
		}),
		cfg: &cfg,
		log: log,
		bot: bot,
	}

	m := pat.New()

	lmt := tollbooth.NewLimiter(20, nil) //20 requests per second
	lmt.SetMessage("Too many requests.")

	//--- our webhook service
	m.Post("/gitlab/webhook", tollbooth.LimitFuncHandler(lmt, webhook.Webhook))

	//--- prometheus metrics
	m.Get("/metrics", promhttp.Handler())
	m.Post("/metrics", promhttp.Handler())

	//--- standard http endpoints
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
	io.WriteString(w, "DayZBR Gitlab Webhooks Service.")
}
func Health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Healthy!\n")
}
func NotFound(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Endpoint not found!")
}

type PushWebhook struct {
	Repository struct {
		Homepage string `json:"homepage"`
	} `json:"repository"`
	PreviousCommit string `json:"before"`
	Commit         string `json:"after"`
	Commits        []struct {
		Message string `json:"message"`
	} `json:"commits"`
}
type IssueWebhook struct {
}

func (h *GitlabWebhook) Webhook(w http.ResponseWriter, r *http.Request) {

	//disable caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	//verify secret token
	token := r.Header.Get("X-Gitlab-Token")
	if token != h.cfg.Gitlab.Token {
		io.WriteString(w, "invalid")
		return
	}

	//increment metrics
	h.total_requests.Inc()
	event := r.Header.Get("X-Gitlab-Event")

	switch event {
	case "Push Hook":
		var push PushWebhook
		err := json.NewDecoder(r.Body).Decode(&push)
		if err != nil {
			h.log.Errorln("Failed to decode push event", err)
			io.WriteString(w, "invalid")
			return
		}

		compare_url := push.Repository.Homepage
		oldest_commit := push.PreviousCommit[0:6]
		newest_commit := push.Commit[0:6]
		compare_url += "/-/compare/" + oldest_commit + "..." + newest_commit

		message := "Pushed " + strconv.Itoa(len(push.Commits)) + " commits to Gitlab. Compare at " + compare_url

		h.bot.MessageChannel(h.cfg.Discord.StatusChannel, message)
		io.WriteString(w, "ok")
	case "Issue Hook":
		var issue IssueWebhook
		err := json.NewDecoder(r.Body).Decode(&issue)
		if err != nil {
			h.log.Errorln("Failed to decode issue event", err)
			io.WriteString(w, "invalid")
			return
		}
		io.WriteString(w, "ok")
	default:
		h.log.Debugln("unhandled event", event)
		io.WriteString(w, "invalid")
	}
}
