package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Logger = defaultLogger()
)

type Config interface {
	LogWriter() io.Writer
	WriteLevel() string
	LogDisableTimestamps() bool
	LogPrettyPrint() bool
	LogFields() map[string]interface{}
}

func defaultLogger() *l {
	log := logrus.StandardLogger()

	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)

	return &l{Entry: logrus.NewEntry(log)}
}

func ConfigureBaseLogger(c Config) error {
	log := logrus.StandardLogger()

	log.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: c.LogDisableTimestamps(),
		PrettyPrint:      c.LogPrettyPrint(),
	})

	log.SetReportCaller(true)

	setOutput(c, log)

	if c.WriteLevel() == "" {
		return nil
	}

	level, err := logrus.ParseLevel(c.WriteLevel())
	if err != nil {
		return err
	}

	log.SetLevel(level)
	entry := logrus.NewEntry(log)
	fields := c.LogFields()
	if fields != nil && len(fields) != 0 {
		entry = log.WithFields(fields)
	}

	Logger = &l{Entry: entry}
	return nil
}

func setOutput(c Config, log *logrus.Logger) {
	if c.LogWriter() != nil {
		log.SetOutput(c.LogWriter())
		return
	}

	if c.WriteLevel() == "" {
		log.SetOutput(io.Discard)
		return
	}

	log.SetOutput(os.Stderr)
}

type l struct {
	*logrus.Entry
}
