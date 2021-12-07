package ui

import (
	"context"
	"time"

	"github.com/mitchellh/go-glint"
	"github.com/tj/go-spin"
)

// Spinner creates a new spinner. The created spinner should NOT be started
// or data races will occur that can result in a panic.
func Spinner() *SpinnerComponent {
	// Create our spinner and setup our default frames
	s := spin.New()
	s.Set(spin.Default)

	return &SpinnerComponent{
		s: s,
		msg: "",
		isSuccess: false,
		isFail: false,
	}
}

type SpinnerComponent struct {
	s    *spin.Spinner
	msg string
	last time.Time
	isSuccess bool
	isFail bool
}

func (c *SpinnerComponent) Body(context.Context) glint.Component {
	var icon string
	
	if (c.isSuccess) {
		icon = "✔"
	} else if (c.isFail) {
		icon = "✖"
	} else {
		current := time.Now()
		if c.last.IsZero() || current.Sub(c.last) > 150*time.Millisecond {
			c.last = current
			c.s.Next()
		}
		
		icon = c.s.Current()
	}

	return glint.Layout(
		glint.Style(glint.Text(icon), glint.Color("cyan")),
		glint.Layout(glint.Text(c.msg)).MarginLeft(1),
	).Row()
}

func (c *SpinnerComponent) SetMessage(msg string) {
	c.msg = msg
}

func (c *SpinnerComponent) Success() {
	c.isSuccess = true
}

func (c *SpinnerComponent) Fail() {
	c.isFail = true
}
