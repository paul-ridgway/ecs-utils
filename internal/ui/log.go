// Package ui holds the terminal output helpers and interactive prompts.
package ui

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	appName   = color.New(color.ReverseVideo).Sprint("ECS Utils ")
	infoBadge = color.New(color.BgHiGreen, color.FgBlack).Sprint(" I ")
	warnBadge = color.New(color.BgHiYellow, color.FgBlack).Sprint(" W ")
	errBadge  = color.New(color.BgHiRed, color.FgBlack).Sprint(" E ")
	infoText  = color.New(color.FgHiGreen)
	warnText  = color.New(color.FgHiYellow)
	errText   = color.New(color.FgHiRed)
)

// Info prints an informational message to stdout.
func Info(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "%s%s %s\n", appName, infoBadge, infoText.Sprintf(format, a...))
}

// Warn prints a warning to stdout.
func Warn(format string, a ...any) {
	fmt.Fprintf(os.Stdout, "%s%s %s\n", appName, warnBadge, warnText.Sprintf(format, a...))
}

// Error prints an error to stderr.
func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s%s %s\n", appName, errBadge, errText.Sprintf(format, a...))
}
