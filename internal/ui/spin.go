package ui

import (
	"fmt"
	"os"
	"runtime"
	"time"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/rs/zerolog"
)

type SpinnerOptions struct {
	Logger  *zerolog.Logger
	NoColor bool
	Indent  string
}

type Spinner struct {
	spinner *spinnerpkg.Spinner
	msg     string
	opts    SpinnerOptions
	logger  *zerolog.Logger
}

func NewSpinner(msg string, opts SpinnerOptions) *Spinner {
	spinnerCharNumb := 14
	if runtime.GOOS == "windows" {
		spinnerCharNumb = 9
	}
	s := &Spinner{
		spinner: spinnerpkg.New(spinnerpkg.CharSets[spinnerCharNumb], 100*time.Millisecond, spinnerpkg.WithWriter(os.Stderr)),
		msg:     msg,
		opts:    opts,
		logger:  opts.Logger,
	}

	if s.logger.Info().Enabled() {
		s.logger.Info().Msgf("Starting: %s", msg)
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
	if s.logger.Info().Enabled() {
		s.logger.Error().Msgf("Failed: %s", s.msg)
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
	if s.logger.Info().Enabled() {
		s.logger.Info().Msgf("Completed: %s", s.msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %s\n",
			s.opts.Indent,
			PrimaryString("✔"),
			s.msg,
		)
	}
}
