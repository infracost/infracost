package apiclient

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// LeveledLogger is a wrapper around logrus.Entry that implements the
// retryablehttp.LeveledLogger interface.
type LeveledLogger struct {
	Logger zerolog.Logger
}

func (l *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error().Msg(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info().Msg(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug().Msg(joinMessage(msg, keysAndValues))
}

func (l *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn().Msg(joinMessage(msg, keysAndValues))
}

func joinMessage(msg string, keysAndValues []interface{}) string {
	s := []string{msg}
	for _, v := range keysAndValues {
		s = append(s, fmt.Sprintf("%v", v))
	}
	return strings.Join(s, " ")
}
