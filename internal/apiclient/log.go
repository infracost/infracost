package apiclient

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LeveledLogger is a wrapper around logrus.Entry that implements the
// retryablehttp.LeveledLogger interface.
type LeveledLogger struct {
	Logger *log.Entry
}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(joinMessage(msg, keysAndValues))
}

func joinMessage(msg string, keysAndValues []interface{}) string {
	s := []string{msg}
	for _, v := range keysAndValues {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return strings.Join(s, " ")
}
