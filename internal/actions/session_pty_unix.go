//go:build !windows

package actions

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var errNoStatusBar = errors.New("status bar needs an interactive terminal")

// terminalIO is the pair of files the session is attached to. In normal use
// it is stdin and stdout; tests pass the slave side of their own pty.
type terminalIO struct {
	in  *os.File
	out *os.File
}

var cursorReport = regexp.MustCompile(`\x1b\[(\d+);\d+R`)

// runWithStatusBar runs cmd on a pseudo-terminal one row shorter than the real
// terminal and keeps a status bar on the spare bottom row. It returns the
// child's exit code.
func runWithStatusBar(cmd *exec.Cmd, tty terminalIO, segments []string) (int, error) {
	inFd, outFd := int(tty.in.Fd()), int(tty.out.Fd())
	if !term.IsTerminal(inFd) || !term.IsTerminal(outFd) {
		return -1, errNoStatusBar
	}
	cols, rows, err := term.GetSize(outFd)
	if err != nil || rows < 3 || cols < 10 {
		return -1, errNoStatusBar
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows - 1), Cols: uint16(cols)})
	if err != nil {
		return -1, err
	}
	defer ptmx.Close()

	oldState, err := term.MakeRaw(inFd)
	if err != nil {
		killGroup(cmd, syscall.SIGKILL)
		_ = cmd.Wait()
		return -1, err
	}
	defer term.Restore(inFd, oldState)

	// A single goroutine owns reads from the terminal; everything else takes
	// keystrokes from this channel.
	keys := make(chan []byte, 16)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := tty.in.Read(buf)
			if n > 0 {
				keys <- append([]byte(nil), buf[:n]...)
			}
			if err != nil {
				close(keys)
				return
			}
		}
	}()

	var outMu sync.Mutex // serialises writes to the terminal
	bar := newStatusBar(rows, cols, segments)

	// Find out where the cursor is so existing output stays put. If it is on
	// the last row, scroll once so that row is free for the footer.
	cursorRow, ok := queryCursorRow(tty, keys, time.Second)
	if !ok || cursorRow >= rows {
		fmt.Fprint(tty.out, "\n")
		cursorRow = rows - 1
	}
	bar.layout(tty.out, cursorRow)

	// Child output -> terminal, repainting the footer whenever the child sends
	// escape sequences that might have cleared it.
	ptyClosed := make(chan struct{})
	go func() {
		defer close(ptyClosed)
		buf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				outMu.Lock()
				_, _ = tty.out.Write(buf[:n])
				if bytes.IndexByte(buf[:n], 0x1b) >= 0 {
					bar.reapply(tty.out)
				}
				outMu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)

	var lastInterrupt time.Time
	var forceKill *time.Timer
	terminating := false

	finish := func() (int, error) {
		if forceKill != nil {
			forceKill.Stop()
		}
		outMu.Lock()
		bar.teardown(tty.out)
		outMu.Unlock()
		return cmd.ProcessState.ExitCode(), nil
	}

	for {
		select {
		case b, open := <-keys:
			if !open {
				keys = nil
				continue
			}
			if bytes.IndexByte(b, 0x03) >= 0 && !terminating {
				now := time.Now()
				if now.Sub(lastInterrupt) < time.Second {
					terminating = true
					outMu.Lock()
					bar.setHint(tty.out, "terminating session...")
					outMu.Unlock()
					killGroup(cmd, syscall.SIGTERM)
					forceKill = time.AfterFunc(3*time.Second, func() { killGroup(cmd, syscall.SIGKILL) })
					continue
				}
				lastInterrupt = now
			}
			_, _ = ptmx.Write(b)

		case <-winch:
			if cols, rows, err := term.GetSize(outFd); err == nil && rows >= 3 && cols >= 10 {
				_ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(rows - 1), Cols: uint16(cols)})
				outMu.Lock()
				bar.resize(tty.out, rows, cols)
				outMu.Unlock()
			}

		case <-ptyClosed:
			ptyClosed = nil
			select {
			case <-waited:
			case <-time.After(2 * time.Second):
				killGroup(cmd, syscall.SIGKILL)
				<-waited
			}
			return finish()

		case <-waited:
			waited = nil
			// Give any remaining output a moment to drain, then stop reading.
			select {
			case <-ptyClosed:
			case <-time.After(200 * time.Millisecond):
				_ = ptmx.Close()
				<-ptyClosed
			}
			return finish()
		}
	}
}

// queryCursorRow asks the terminal for the cursor position and parses the
// reply from the key stream. Typed-ahead bytes arriving with it are dropped.
func queryCursorRow(tty terminalIO, keys <-chan []byte, timeout time.Duration) (int, bool) {
	if _, err := fmt.Fprint(tty.out, "\x1b[6n"); err != nil {
		return 0, false
	}
	var buf []byte
	deadline := time.After(timeout)
	for {
		select {
		case b, open := <-keys:
			if !open {
				return 0, false
			}
			buf = append(buf, b...)
			if m := cursorReport.FindSubmatch(buf); m != nil {
				row, err := strconv.Atoi(string(m[1]))
				return row, err == nil
			}
		case <-deadline:
			return 0, false
		}
	}
}

// killGroup signals the child's whole process group. pty.Start makes the child
// a session leader, so its group also holds the session-manager-plugin.
func killGroup(cmd *exec.Cmd, sig syscall.Signal) {
	if cmd.Process == nil {
		return
	}
	if err := syscall.Kill(-cmd.Process.Pid, sig); err != nil {
		_ = cmd.Process.Signal(sig)
	}
}
