package ui

import (
	"fmt"
	"os"
	"runtime"
	"time"

	spinnerpkg "github.com/briandowns/spinner"
	log "github.com/sirupsen/logrus"
)

type SpinnerOptions struct {
	EnableLogging bool
	NoColor       bool
	Indent        string
}

type Spinner struct {
	spinner *spinnerpkg.Spinner
	msg     string
	opts    SpinnerOptions
}

// SpinnerFunc defines a function that returns a Spinner which can be used
// to report the progress of a certain action.
type SpinnerFunc func(msg string) *Spinner

func NewSpinner(msg string, opts SpinnerOptions) *Spinner {
	spinnerCharNumb := 14
	if runtime.GOOS == "windows" {
		spinnerCharNumb = 9
	}
	s := &Spinner{
		spinner: spinnerpkg.New(spinnerpkg.CharSets[spinnerCharNumb], 100*time.Millisecond, spinnerpkg.WithWriter(os.Stderr)),
		msg:     msg,
		opts:    opts,
	}

	if s.opts.EnableLogging {
		log.Infof("Starting: %s", msg)
	} else {
		s.spinner.Prefix = opts.Indent
		s.spinner.Suffix = fmt.Sprintf(" %s", msg)
		if !s.opts.NoColor {
			_ = s.spinner.Color("fgHiCyan", "bold")
		}
		s.spinner.Start()
	}

	return s
}

func (s *Spinner) Stop() {
	s.spinner.Stop()
}

func (s *Spinner) Fail() {
	if s.spinner == nil || !s.spinner.Active() {
		return
	}
	s.Stop()
	if s.opts.EnableLogging {
		log.Errorf("Failed: %s", s.msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %s\n",
			s.opts.Indent,
			ErrorString("✖"),
			s.msg,
		)
	}
}

func (s *Spinner) SuccessWithMessage(newMsg string) {
	s.msg = newMsg
	s.Success()
}

func (s *Spinner) Success() {
	if !s.spinner.Active() {
		return
	}
	s.Stop()
	if s.opts.EnableLogging {
		log.Infof("Completed: %s", s.msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %s\n",
			s.opts.Indent,
			PrimaryString("✔"),
			s.msg,
		)
	}
}
