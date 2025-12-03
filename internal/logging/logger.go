package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Logger is the global logger for the logging package. Callers should use this Logger
	// to ensure that logging functionality is consistent across the Infracost codebase.
	//
	// It is advised to create child Loggers as needed and pass them into packages with
	// relevant log metadata. This can be done by using the With method:
	//
	//		childLogger := logging.Logger.With().Str("additional", "field")
	//		foo := MyStruct{Logger: childLogger}
	//		foo.DoSomething()
	//
	// Child loggers will inherit the parent metadata fields, unless the child logger sets metadata
	// field information with the same key. In this case child fields will overwrite the parent field.
	Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

// Config is an interface that fetches Logger configuration details for the Logger.
// This is used so that the logging packages does not use any other Infracost packages,
// avoiding circular deps problems.
type Config interface {
	// LogWriter returns the writer the Logger should use to write logs to.
	// In most cases this should be stderr, but it can also be a file.
	LogWriter() io.Writer
	// WriteLevel is the log level that the Logger writes to LogWriter.
	WriteLevel() string
	// LogFields sets the meta fields that are added to any log line entries.
	LogFields() map[string]any
}

// ConfigureBaseLogger configures the global Logger using the provided Config.
func ConfigureBaseLogger(c Config) error {
	setOutput(c)

	level, err := zerolog.ParseLevel(c.WriteLevel())
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = log.Level(level)

	Logger = log.Logger
	return nil
}

func setOutput(c Config) {
	log.Logger = log.Output(c.LogWriter())
}
