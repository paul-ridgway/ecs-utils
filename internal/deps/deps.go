// Package deps checks for the external tools ecs-utils relies on (the AWS CLI
// and the Session Manager plugin) and installs them when they are missing.
package deps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

type dependency struct {
	name        string
	binary      string
	versionArgs []string
	extraDirs   []string
	install     func(ctx context.Context) error
}

func awsCLI() dependency {
	return dependency{
		name:        "AWS CLI",
		binary:      "aws",
		versionArgs: []string{"--version"},
		extraDirs:   awsCLIDirs(),
		install:     installAWSCLI,
	}
}

func sessionManagerPlugin() dependency {
	return dependency{
		name:        "Session Manager plugin",
		binary:      "session-manager-plugin",
		versionArgs: []string{"--version"},
		extraDirs:   sessionManagerPluginDirs(),
		install:     installSessionManagerPlugin,
	}
}

// Ensure checks that every dependency is available and offers to install the
// ones that are not. With assumeYes the installation runs without asking.
func Ensure(ctx context.Context, assumeYes bool) error {
	var missing []dependency
	for _, dep := range []dependency{awsCLI(), sessionManagerPlugin()} {
		if path, ok := dep.locate(); ok {
			ui.Info("Found %s %s (%s)", dep.name, dep.version(ctx, path), path)
			continue
		}
		ui.Warn("%s (%s) was not found", dep.name, dep.binary)
		missing = append(missing, dep)
	}
	if len(missing) == 0 {
		return nil
	}

	install := assumeYes
	if !install {
		var err error
		install, err = ui.Confirm("Install the missing dependencies now?", true)
		if err != nil {
			return err
		}
	}
	if !install {
		ui.Warn("Continuing without them; 'Execute Command' needs both the AWS CLI and the Session Manager plugin")
		return nil
	}

	for _, dep := range missing {
		ui.Info("Installing %s...", dep.name)
		if err := dep.install(ctx); err != nil {
			return fmt.Errorf("installing %s: %w", dep.name, err)
		}
		path, ok := dep.locate()
		if !ok {
			return fmt.Errorf("%s was installed but %s still cannot be found; open a new shell and try again", dep.name, dep.binary)
		}
		ui.Info("Installed %s %s (%s)", dep.name, dep.version(ctx, path), path)
	}
	return nil
}

// locate finds the binary on PATH or in the well-known install locations. A
// hit outside PATH is added to this process's PATH so child processes see it.
func (d dependency) locate() (string, bool) {
	if path, err := exec.LookPath(d.binary); err == nil {
		return path, true
	}
	for _, dir := range d.extraDirs {
		path := filepath.Join(dir, d.binary+exeSuffix())
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			ui.Warn("%s found at %s, which is not on your PATH; add %s to PATH to use it from your shell", d.binary, path, dir)
			prependPath(dir)
			return path, true
		}
	}
	return "", false
}

func (d dependency) version(ctx context.Context, path string) string {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, d.versionArgs...).CombinedOutput()
	if err != nil {
		return "(version unknown)"
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return strings.TrimSpace(line)
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func prependPath(dir string) {
	current := os.Getenv("PATH")
	for _, p := range filepath.SplitList(current) {
		if p == dir {
			return
		}
	}
	_ = os.Setenv("PATH", dir+string(os.PathListSeparator)+current)
}

// localBinDir is where user-local installs put their launchers (~/.local/bin).
func localBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "bin")
}

// ensureLocalBinOnPath warns when ~/.local/bin is not on PATH and adds it for
// the rest of this run.
func ensureLocalBinOnPath() {
	dir := localBinDir()
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == dir {
			return
		}
	}
	ui.Warn("%s is not on your PATH. Add it to your shell profile, e.g.:", dir)
	ui.Warn(`    export PATH="$HOME/.local/bin:$PATH"`)
	prependPath(dir)
}

// download fetches url to dest, showing progress when the size is known.
func download(ctx context.Context, url, dest string) error {
	ui.Info("Downloading %s", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %s fetching %s", resp.Status, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	progress := &progressWriter{total: resp.ContentLength}
	if _, err := io.Copy(io.MultiWriter(f, progress), resp.Body); err != nil {
		return err
	}
	progress.finish()
	return nil
}

type progressWriter struct {
	total       int64
	written     int64
	lastPercent int64
}

func (p *progressWriter) Write(b []byte) (int, error) {
	p.written += int64(len(b))
	if p.total > 0 {
		if percent := p.written * 100 / p.total; percent/5 > p.lastPercent/5 {
			p.lastPercent = percent
			fmt.Fprintf(os.Stdout, "\r  %3d%% of %s", percent, humanSize(p.total))
		}
	}
	return len(b), nil
}

func (p *progressWriter) finish() {
	if p.total > 0 {
		if p.lastPercent < 100 {
			fmt.Fprintf(os.Stdout, "\r  100%% of %s", humanSize(p.total))
		}
		fmt.Fprintln(os.Stdout)
	} else {
		fmt.Fprintf(os.Stdout, "  downloaded %s\n", humanSize(p.written))
	}
}

func humanSize(n int64) string {
	const mb = 1024 * 1024
	if n >= mb {
		return fmt.Sprintf("%.1f MB", float64(n)/mb)
	}
	return fmt.Sprintf("%.0f KB", float64(n)/1024)
}

// runCmd runs a command with the terminal attached, echoing what it runs.
func runCmd(ctx context.Context, name string, args ...string) error {
	ui.Info("Running: %s %s", name, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// replaceSymlink points link at target, replacing whatever was there before.
func replaceSymlink(target, link string) error {
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	if _, err := os.Lstat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return err
		}
	}
	return os.Symlink(target, link)
}
