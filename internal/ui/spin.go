package ui

import (
	"fmt"
	"os"
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

func NewSpinner(msg string, opts SpinnerOptions) *Spinner {
	s := &Spinner{
		spinner: spinnerpkg.New(spinnerpkg.CharSets[14], 100*time.Millisecond, spinnerpkg.WithWriter(os.Stderr)),
		msg:     msg,
		opts:    opts,
	}

	if s.opts.EnableLogging {
		log.Infof("starting: %s", msg)
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
	if !s.spinner.Active() {
		return
	}
	s.spinner.Stop()
	if s.opts.EnableLogging {
		log.Errorf("failed: %s", s.msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %s\n",
			s.opts.Indent,
			ErrorString("✖"),
			s.msg,
		)
	}
}

func (s *Spinner) Success() {
	if !s.spinner.Active() {
		return
	}
	s.spinner.Stop()
	if s.opts.EnableLogging {
		log.Infof("completed: %s", s.msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %s\n",
			s.opts.Indent,
			PrimaryString("✔"),
			s.msg,
		)
	}
}
