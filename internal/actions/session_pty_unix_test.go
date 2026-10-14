//go:build !windows

package actions

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
)

// harness opens a pty pair, replies to cursor-position queries on the master
// side and records everything the session writes to the terminal.
type harness struct {
	t      *testing.T
	master *os.File
	tty    terminalIO
	mu     sync.Mutex
	out    bytes.Buffer
	done   chan struct{}
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	if err := pty.Setsize(master, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, master: master, tty: terminalIO{in: slave, out: slave}, done: make(chan struct{})}
	t.Cleanup(func() { slave.Close(); master.Close() })
	go func() {
		defer close(h.done)
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				h.mu.Lock()
				h.out.Write(buf[:n])
				h.mu.Unlock()
				if bytes.Contains(buf[:n], []byte("\x1b[6n")) {
					_, _ = master.Write([]byte("\x1b[24;1R")) // cursor on the last row
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return h
}

func (h *harness) output() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.out.String()
}

func (h *harness) waitFor(substr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(h.output(), substr) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestRunWithStatusBarShowsFooterAndExitCode(t *testing.T) {
	h := newHarness(t)
	cmd := exec.Command("sh", "-c", `printf 'hello from child\n'; printf '\033[2J'; exit 3`)
	code, err := runWithStatusBar(cmd, h.tty, []string{"AWS SSM SESSION", "cluster: demo"})
	if err != nil {
		t.Fatal(err)
	}
	if code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
	if !h.waitFor("\x1b[r", 2*time.Second) {
		t.Fatal("teardown sequence never reached the terminal")
	}
	out := h.output()
	for _, want := range []string{
		"\x1b[1;23r",                  // scroll region leaves the last row free
		"\x1b[23;1H",                  // cursor moved up after scrolling the last row
		"hello from child",            // child output passed through
		"AWS SSM SESSION  │  cluster", // footer text
		"\x1b[24;1H",                  // footer drawn on the last row
		"\x1b[r",                      // scroll region reset on teardown
	} {
		if !strings.Contains(out, want) {
			t.Errorf("terminal output missing %q", want)
		}
	}
	// The footer must be repainted after the child cleared the screen.
	if strings.Count(out, "AWS SSM SESSION") < 2 {
		t.Errorf("expected the footer to be redrawn after \\x1b[2J, got %d draws", strings.Count(out, "AWS SSM SESSION"))
	}
}

func TestRunWithStatusBarDoubleInterruptTerminates(t *testing.T) {
	h := newHarness(t)
	cmd := exec.Command("sh", "-c", `trap '' INT; sleep 30`)
	result := make(chan int, 1)
	go func() {
		code, _ := runWithStatusBar(cmd, h.tty, []string{"AWS SSM SESSION"})
		result <- code
	}()
	if !h.waitFor("ctrl-c twice to exit", 5*time.Second) {
		t.Fatal("status bar never appeared")
	}
	// One ctrl-c is forwarded to the child (which ignores it); a second one
	// within a second ends the session.
	_, _ = h.master.Write([]byte{3})
	time.Sleep(100 * time.Millisecond)
	_, _ = h.master.Write([]byte{3})
	select {
	case <-result:
	case <-time.After(10 * time.Second):
		t.Fatal("session did not terminate after a double ctrl-c")
	}
	if !strings.Contains(h.output(), "terminating session...") {
		t.Error("expected the footer hint to change while terminating")
	}
}

func TestStatusBarTextFitsWidth(t *testing.T) {
	bar := newStatusBar(24, 30, []string{"AWS SSM SESSION", "cluster: a-rather-long-cluster-name", "task: abc"})
	text := bar.text()
	if n := len([]rune(text)); n != 30 {
		t.Errorf("rendered %d cells, want 30", n)
	}
	if !strings.HasPrefix(text, " AWS SSM SESSION") {
		t.Errorf("unexpected text %q", text)
	}
	wide := newStatusBar(24, 120, []string{"AWS SSM SESSION", "cluster: demo"})
	if got := wide.text(); !strings.Contains(got, "ctrl-c twice to exit") || len([]rune(got)) != 120 {
		t.Errorf("unexpected wide text %q", got)
	}
}
