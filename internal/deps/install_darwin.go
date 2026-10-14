//go:build darwin

package deps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func awsCLIDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{localBinDir(), "/usr/local/bin", "/opt/homebrew/bin", filepath.Join(home, "aws-cli")}
}

func sessionManagerPluginDirs() []string {
	return []string{localBinDir(), "/usr/local/bin", "/opt/homebrew/bin", "/usr/local/sessionmanagerplugin/bin"}
}

// installAWSCLI prefers Homebrew. Without it, the official .pkg is installed
// into the user's home directory (no root needed) and linked into ~/.local/bin.
func installAWSCLI(ctx context.Context) error {
	if brew, err := exec.LookPath("brew"); err == nil {
		return runCmd(ctx, brew, "install", "awscli")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "ecs-utils-awscli-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	pkgPath := filepath.Join(tmp, "AWSCLIV2.pkg")
	if err := download(ctx, "https://awscli.amazonaws.com/AWSCLIV2.pkg", pkgPath); err != nil {
		return err
	}
	choices := filepath.Join(tmp, "choices.xml")
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <array>
    <dict>
      <key>choiceAttribute</key>
      <string>customLocation</string>
      <key>attributeSetting</key>
      <string>%s</string>
      <key>choiceIdentifier</key>
      <string>default</string>
    </dict>
  </array>
</plist>
`, home)
	if err := os.WriteFile(choices, []byte(xml), 0o644); err != nil {
		return err
	}
	if err := runCmd(ctx, "installer", "-pkg", pkgPath, "-target", "CurrentUserHomeDirectory", "-applyChoiceChangesXML", choices); err != nil {
		return err
	}
	for _, name := range []string{"aws", "aws_completer"} {
		if err := replaceSymlink(filepath.Join(home, "aws-cli", name), filepath.Join(localBinDir(), name)); err != nil {
			return err
		}
	}
	ensureLocalBinOnPath()
	return nil
}

// installSessionManagerPlugin prefers Homebrew. Without it, the official
// bundle is unpacked into ~/.local/sessionmanagerplugin and linked into
// ~/.local/bin.
func installSessionManagerPlugin(ctx context.Context) error {
	if brew, err := exec.LookPath("brew"); err == nil {
		return runCmd(ctx, brew, "install", "--cask", "session-manager-plugin")
	}

	flavour := map[string]string{"amd64": "mac_64bit", "arm64": "mac_arm64"}[runtime.GOARCH]
	if flavour == "" {
		return fmt.Errorf("no Session Manager plugin build is available for darwin/%s", runtime.GOARCH)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "ecs-utils-smp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	zipPath := filepath.Join(tmp, "sessionmanager-bundle.zip")
	url := fmt.Sprintf("https://s3.amazonaws.com/session-manager-downloads/plugin/latest/%s/sessionmanager-bundle.zip", flavour)
	if err := download(ctx, url, zipPath); err != nil {
		return err
	}

	installDir := filepath.Join(home, ".local", "sessionmanagerplugin")
	if err := extractZip(zipPath, installDir, "sessionmanager-bundle/"); err != nil {
		return err
	}
	binary := filepath.Join(installDir, "bin", "session-manager-plugin")
	if err := os.Chmod(binary, 0o755); err != nil {
		return err
	}
	if err := replaceSymlink(binary, filepath.Join(localBinDir(), "session-manager-plugin")); err != nil {
		return err
	}
	ensureLocalBinOnPath()
	return nil
}
