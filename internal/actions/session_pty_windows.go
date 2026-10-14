//go:build windows

package actions

import (
	"errors"
	"os"
	"os/exec"
)

var errNoStatusBar = errors.New("status bar is not supported on Windows")

type terminalIO struct {
	in  *os.File
	out *os.File
}

func runWithStatusBar(_ *exec.Cmd, _ terminalIO, _ []string) (int, error) {
	return -1, errNoStatusBar
}
