package actions

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// runPlain runs the session with the terminal handed straight to the child.
// It is the fallback when a status bar cannot be shown (Windows, or no tty).
func runPlain(cmd *exec.Cmd) (int, error) {
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// Trap ctrl-c: a single press is passed through to the container (the
	// child shares our process group so it receives the signal too); two
	// presses within a second end the session.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)

	if err := cmd.Start(); err != nil {
		return -1, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	var last time.Time
	for {
		select {
		case <-interrupts:
			now := time.Now()
			if now.Sub(last) < time.Second {
				fmt.Println()
				ui.Warn("Terminating session...")
				terminateSession(cmd)
				select {
				case <-done:
				case <-time.After(3 * time.Second):
					_ = cmd.Process.Kill()
					<-done
				}
				return cmd.ProcessState.ExitCode(), nil
			}
			last = now
		case err := <-done:
			var exitErr *exec.ExitError
			if err != nil && !errors.As(err, &exitErr) {
				return -1, err
			}
			return cmd.ProcessState.ExitCode(), nil
		}
	}
}
