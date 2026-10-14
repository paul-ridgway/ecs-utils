package actions

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// statusBar is the persistent footer shown while a session is running. It
// lives on the last row of the terminal, outside the scroll region that the
// child process writes into.
type statusBar struct {
	mu       sync.Mutex
	rows     int
	cols     int
	segments []string
	hint     string
}

const (
	barStyle = "\x1b[0;30;48;5;214m" // black on AWS orange
	barReset = "\x1b[0m"
)

func newStatusBar(rows, cols int, segments []string) *statusBar {
	return &statusBar{rows: rows, cols: cols, segments: segments, hint: "ctrl-c twice to exit"}
}

// layout sets the scroll region to everything above the footer, places the
// cursor on cursorRow and draws the bar. It is used once at start-up.
func (b *statusBar) layout(w io.Writer, cursorRow int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Fprintf(w, "\x1b[1;%dr\x1b[%d;1H", b.rows-1, cursorRow)
	b.drawLocked(w)
}

// draw repaints the footer without disturbing the cursor or attributes.
func (b *statusBar) draw(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.drawLocked(w)
}

// reapply re-asserts the scroll region (which a child may have reset) and
// repaints the footer.
func (b *statusBar) reapply(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Fprintf(w, "\x1b7\x1b[1;%dr\x1b8", b.rows-1)
	b.drawLocked(w)
}

// resize records a new terminal size and repaints.
func (b *statusBar) resize(w io.Writer, rows, cols int) {
	b.mu.Lock()
	b.rows, b.cols = rows, cols
	b.mu.Unlock()
	b.reapply(w)
}

// setHint replaces the trailing hint text and repaints.
func (b *statusBar) setHint(w io.Writer, hint string) {
	b.mu.Lock()
	b.hint = hint
	b.mu.Unlock()
	b.draw(w)
}

// teardown removes the scroll region and clears the footer row, leaving the
// cursor at the start of that row so normal output can continue there.
func (b *statusBar) teardown(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fmt.Fprintf(w, "\x1b[r\x1b[%d;1H\x1b[2K", b.rows)
}

func (b *statusBar) drawLocked(w io.Writer) {
	fmt.Fprintf(w, "\x1b7\x1b[%d;1H%s%s%s\x1b8", b.rows, barStyle, b.text(), barReset)
}

// text renders the bar to exactly cols cells, dropping trailing segments when
// the terminal is too narrow for all of them.
func (b *statusBar) text() string {
	segments := append(append([]string{}, b.segments...), b.hint)
	render := func(n int) []rune {
		return []rune(" " + strings.Join(segments[:n], "  │  ") + " ")
	}
	line := render(len(segments))
	for n := len(segments) - 1; len(line) > b.cols && n > 0; n-- {
		line = render(n)
	}
	if len(line) > b.cols {
		line = line[:b.cols]
	}
	return string(line) + strings.Repeat(" ", b.cols-len(line))
}
