package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	// Logger is the global logger for the logging package. Callers should use this Logger
	// to ensure that logging functionality is consistent across the Infracost codebase.
	//
	// It is advised to create child Loggers as needed and pass them into packages with
	// relevant log metadata. This can be done by using the WithFields method:
	//
	//		childLogger := log.Logger.WithFields(logrus.Fields{"contextual_data": "value"})
	//		foo := MyStruct{Logger: childLogger}
	//		foo.DoSomething()
	//
	// Child loggers will inherit the parent metadata fields, unless the child logger sets metadata
	// field information with the same key. In this case child fields will overwrite the parent field.
	Logger = defaultLogger()
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
	// LogDisableTimestamps sets if the log entry contains the timestamp the line is written at.
	LogDisableTimestamps() bool
	// LogPrettyPrint sets if the log entry is JSON pretty printed to the writer.
	LogPrettyPrint() bool
	// LogFields sets the meta fields that are added to any log line entries.
	LogFields() map[string]interface{}
	// ReportCaller sets whether the log entry writes the filename to the log line.
	ReportCaller() bool
}

type l struct {
	*logrus.Entry
}

func defaultLogger() *l {
	log := logrus.StandardLogger()

	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)

	return &l{Entry: logrus.NewEntry(log)}
}

// ConfigureBaseLogger configures the global Logger using the provided Config.
func ConfigureBaseLogger(c Config) error {
	log := logrus.StandardLogger()

	log.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: c.LogDisableTimestamps(),
		PrettyPrint:      c.LogPrettyPrint(),
	})

	log.SetReportCaller(c.ReportCaller())

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
