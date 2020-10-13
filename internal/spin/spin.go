package spin

import (
	"fmt"
	"os"
	"time"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	log "github.com/sirupsen/logrus"
)

type Spinner struct {
	spinner *spinnerpkg.Spinner
	msg     string
}

func NewSpinner(msg string) *Spinner {
	s := &Spinner{
		spinner: spinnerpkg.New(spinnerpkg.CharSets[14], 100*time.Millisecond, spinnerpkg.WithWriter(os.Stderr)),
		msg:     msg,
	}

	if config.Config.IsLogging() {
		log.Infof("starting: %s", msg)
	} else {
		s.spinner.Prefix = "  "
		s.spinner.Suffix = fmt.Sprintf(" %s", msg)
		if !config.Config.NoColor {
			_ = s.spinner.Color("fgHiBlue", "bold")
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
	if config.Config.IsLogging() {
		log.Errorf("failed: %s", s.msg)
	} else {
		fmt.Fprintln(os.Stderr, color.HiRedString("  ✖ %s", s.msg))
	}
}

func (s *Spinner) Success() {
	if !s.spinner.Active() {
		return
	}
	s.spinner.Stop()
	if config.Config.IsLogging() {
		log.Infof("completed: %s", s.msg)
	} else {
		fmt.Fprintln(os.Stderr, color.GreenString("  ✔ %s", s.msg))
	}
}
