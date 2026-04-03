package ui

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/infracost/infracost/internal/config"
)

var primary = color.New(color.FgHiCyan)

var yellow = color.New(color.FgYellow)
var red = color.New(color.FgHiRed)
var green = color.New(color.FgHiGreen)

var bold = color.New(color.Bold)
var faint = color.New(color.Faint)
var underline = color.New(color.Underline)

var primaryLink = color.New(color.Underline).Add(color.Bold)

func PrimaryString(msg string) string {
	return primary.Sprint(msg)
}

func PrimaryStringf(msg string, a ...any) string {
	return PrimaryString(fmt.Sprintf(msg, a...))
}

func SuccessString(msg string) string {
	return green.Sprint(msg)
}

func SuccessStringf(msg string, a ...any) string {
	return SuccessString(fmt.Sprintf(msg, a...))
}

func ErrorString(msg string) string {
	return red.Sprint(msg)
}

func ErrorStringf(msg string, a ...any) string {
	return ErrorString(fmt.Sprintf(msg, a...))
}

func WarningString(msg string) string {
	return yellow.Sprint(msg)
}

func WarningStringf(msg string, a ...any) string {
	return WarningString(fmt.Sprintf(msg, a...))
}

func LinkString(msg string) string {
	return primaryLink.Sprint(msg)
}

func LinkStringf(msg string, a ...any) string {
	return LinkString(fmt.Sprintf(msg, a...))
}

func SecondaryLinkString(msg string) string {
	return underline.Sprint(msg)
}

func SecondaryLinkStringf(msg string, a ...any) string {
	return SecondaryLinkString(fmt.Sprintf(msg, a...))
}

func BoldString(msg string) string {
	return bold.Sprint(msg)
}

func BoldStringf(msg string, a ...any) string {
	return BoldString(fmt.Sprintf(msg, a...))
}

func FaintString(msg string) string {
	return faint.Sprint(msg)
}

func FaintStringf(msg string, a ...any) string {
	return FaintString(fmt.Sprintf(msg, a...))
}

func UnderlineString(msg string) string {
	return underline.Sprint(msg)
}

func UnderlineStringf(msg string, a ...any) string {
	return UnderlineString(fmt.Sprintf(msg, a...))
}

// FormatIfNotCI runs the formatFunc if the current run context is not a CI run.
func FormatIfNotCI(ctx *config.RunContext, formatFunc func(string) string, value string) string {
	if ctx.IsCIRun() {
		return fmt.Sprintf("%q", value)
	}

	return formatFunc(value)
}
