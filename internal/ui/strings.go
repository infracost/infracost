package ui

import (
	"fmt"

	"github.com/fatih/color"
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

func PrimaryStringf(msg string, a ...interface{}) string {
	return PrimaryString(fmt.Sprintf(msg, a...))
}

func SuccessString(msg string) string {
	return green.Sprint(msg)
}

func SuccessStringf(msg string, a ...interface{}) string {
	return SuccessString(fmt.Sprintf(msg, a...))
}

func ErrorString(msg string) string {
	return red.Sprint(msg)
}

func ErrorStringf(msg string, a ...interface{}) string {
	return ErrorString(fmt.Sprintf(msg, a...))
}

func WarningString(msg string) string {
	return yellow.Sprint(msg)
}

func WarningStringf(msg string, a ...interface{}) string {
	return WarningString(fmt.Sprintf(msg, a...))
}

func LinkString(msg string) string {
	return primaryLink.Sprint(msg)
}

func LinkStringf(msg string, a ...interface{}) string {
	return LinkString(fmt.Sprintf(msg, a...))
}

func SecondaryLinkString(msg string) string {
	return underline.Sprint(msg)
}

func SecondaryLinkStringf(msg string, a ...interface{}) string {
	return SecondaryLinkString(fmt.Sprintf(msg, a...))
}

func BoldString(msg string) string {
	return bold.Sprint(msg)
}

func BoldStringf(msg string, a ...interface{}) string {
	return BoldString(fmt.Sprintf(msg, a...))
}

func FaintString(msg string) string {
	return faint.Sprint(msg)
}

func FaintStringf(msg string, a ...interface{}) string {
	return FaintString(fmt.Sprintf(msg, a...))
}

func UnderlineString(msg string) string {
	return underline.Sprint(msg)
}

func UnderlineStringf(msg string, a ...interface{}) string {
	return UnderlineString(fmt.Sprintf(msg, a...))
}
