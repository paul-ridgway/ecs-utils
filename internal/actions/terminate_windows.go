//go:build windows

package actions

import "os/exec"

// terminateSession stops the AWS CLI process. Windows has no process groups
// to signal, so the plugin it spawned is left to exit with it.
func terminateSession(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}
