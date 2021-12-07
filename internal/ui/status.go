package ui

import (
	"context"

	"github.com/mitchellh/go-glint"
)

// Spinner creates a new spinner. The created spinner should NOT be started
// or data races will occur that can result in a panic.
func Status() *StatusComponent {
	// Create our spinner and setup our default frames
	s := ""

	return &StatusComponent{
		s: s,
	}
}

type StatusComponent struct {
	s    string
}

func (c *StatusComponent) Body(context.Context) glint.Component {
	return glint.Text(c.s)
}

func (c *StatusComponent) SetStatus(s string) {
	c.s = s
}
