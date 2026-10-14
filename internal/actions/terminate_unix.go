//go:build !windows

package actions

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// terminateSession asks the AWS CLI and the session-manager-plugin it spawned
// to stop. When we lead the process group (the usual case for an interactive
// job) the whole group is signalled so the plugin does not outlive us.
func terminateSession(cmd *exec.Cmd) {
	if pgrp := syscall.Getpgrp(); pgrp == os.Getpid() {
		signal.Ignore(syscall.SIGTERM)
		_ = syscall.Kill(-pgrp, syscall.SIGTERM)
		return
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
}
