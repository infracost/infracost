package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow)
var red = color.New(color.FgHiRed)
var green = color.New(color.FgHiGreen)
var blue = color.New(color.FgHiBlue)
var magenta = color.New(color.FgHiCyan)

var bold = color.New(color.Bold)
var faded = color.New(color.FgHiBlack)

func PrimaryString(msg string) string {
	return magenta.Sprint(msg)
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

func LinkString(url string) string {
	return blue.Sprint(url)
}

func LinkStringf(msg string, a ...interface{}) string {
	return LinkString(fmt.Sprintf(msg, a...))
}

func BoldString(url string) string {
	return bold.Sprint(url)
}

func BoldStringf(msg string, a ...interface{}) string {
	return BoldString(fmt.Sprintf(msg, a...))
}

func FadedString(url string) string {
	return faded.Sprint(url)
}

func FadedStringf(msg string, a ...interface{}) string {
	return FadedString(fmt.Sprintf(msg, a...))
}
