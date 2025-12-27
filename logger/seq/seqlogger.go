package seq

import (
	"os"

	"github.com/nullseed/logruseq"
	"github.com/sirupsen/logrus"
)

type SeqConfig struct {
	Endpoint string
	Apikey   string
	Appname  string
	Seqlevel uint32
	Loglevel uint32
}

func NewLogger(config *SeqConfig) *logrus.Entry {
	seq_endpoint := config.Endpoint
	seq_apikey := config.Apikey
	seq_appname := config.Appname

	seq_loglevel := config.Seqlevel

	var levels []logrus.Level
	for i := uint32(0); i <= seq_loglevel; i++ {
		levels = append(levels, logrus.Level(i))
	}

	logrus.SetLevel(logrus.Level(config.Loglevel))
	logrus.AddHook(
		logruseq.NewSeqHook(
			seq_endpoint,
			logruseq.OptionAPIKey(seq_apikey),
			logruseq.OptionLevels(levels),
		),
	)

	//pid field
	pid := os.Getpid()
	seqLogger := logrus.WithField("pid", pid)
	//app name field
	seqLogger = seqLogger.WithField("name", seq_appname)
	//hostname field
	host, err := os.Hostname()
	if err == nil {
		seqLogger = seqLogger.WithField("hostname", host)
	}

	return seqLogger
}
