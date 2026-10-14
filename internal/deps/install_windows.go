//go:build windows

package deps

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

func programFiles() string {
	if dir := os.Getenv("ProgramFiles"); dir != "" {
		return dir
	}
	return `C:\Program Files`
}

func awsCLIDirs() []string {
	return []string{filepath.Join(programFiles(), "Amazon", "AWSCLIV2")}
}

func sessionManagerPluginDirs() []string {
	return []string{filepath.Join(programFiles(), "Amazon", "SessionManagerPlugin", "bin")}
}

// installAWSCLI prefers winget and falls back to the official MSI.
func installAWSCLI(ctx context.Context) error {
	if winget, err := exec.LookPath("winget"); err == nil {
		return runCmd(ctx, winget, "install", "--id", "Amazon.AWSCLI", "-e", "--accept-package-agreements", "--accept-source-agreements")
	}
	tmp, err := os.MkdirTemp("", "ecs-utils-awscli-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	msiPath := filepath.Join(tmp, "AWSCLIV2.msi")
	if err := download(ctx, "https://awscli.amazonaws.com/AWSCLIV2.msi", msiPath); err != nil {
		return err
	}
	if err := runCmd(ctx, "msiexec", "/i", msiPath, "/passive"); err != nil {
		return err
	}
	ui.Warn("Open a new terminal after installing so the updated PATH is picked up")
	return nil
}

// installSessionManagerPlugin prefers winget and falls back to the official
// installer, which runs interactively.
func installSessionManagerPlugin(ctx context.Context) error {
	if winget, err := exec.LookPath("winget"); err == nil {
		return runCmd(ctx, winget, "install", "--id", "Amazon.SessionManagerPlugin", "-e", "--accept-package-agreements", "--accept-source-agreements")
	}
	tmp, err := os.MkdirTemp("", "ecs-utils-smp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	setupPath := filepath.Join(tmp, "SessionManagerPluginSetup.exe")
	if err := download(ctx, "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/windows/SessionManagerPluginSetup.exe", setupPath); err != nil {
		return err
	}
	if err := runCmd(ctx, setupPath); err != nil {
		return err
	}
	ui.Warn("Open a new terminal after installing so the updated PATH is picked up")
	return nil
}
