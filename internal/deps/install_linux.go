//go:build linux

package deps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func awsCLIDirs() []string {
	return []string{localBinDir(), "/usr/local/bin"}
}

func sessionManagerPluginDirs() []string {
	return []string{localBinDir(), "/usr/local/sessionmanagerplugin/bin"}
}

// installAWSCLI installs AWS CLI v2 from the official bundle into
// ~/.local/aws-cli with a launcher in ~/.local/bin, so no root is needed.
func installAWSCLI(ctx context.Context) error {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	if arch == "" {
		return fmt.Errorf("no AWS CLI build is available for linux/%s", runtime.GOARCH)
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

	zipPath := filepath.Join(tmp, "awscliv2.zip")
	if err := download(ctx, fmt.Sprintf("https://awscli.amazonaws.com/awscli-exe-linux-%s.zip", arch), zipPath); err != nil {
		return err
	}
	if err := extractZip(zipPath, tmp, ""); err != nil {
		return err
	}

	installDir := filepath.Join(home, ".local", "aws-cli")
	binDir := localBinDir()
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	installer := filepath.Join(tmp, "aws", "install")
	if err := os.Chmod(installer, 0o755); err != nil {
		return err
	}
	args := []string{"--install-dir", installDir, "--bin-dir", binDir}
	if _, err := os.Stat(installDir); err == nil {
		args = append(args, "--update")
	}
	if err := runCmd(ctx, installer, args...); err != nil {
		return err
	}
	ensureLocalBinOnPath()
	return nil
}

// installSessionManagerPlugin unpacks the official .deb (which only contains a
// self-contained binary) into ~/.local/sessionmanagerplugin and links it into
// ~/.local/bin. This works on any distribution and needs no root.
func installSessionManagerPlugin(ctx context.Context) error {
	flavour := map[string]string{"amd64": "ubuntu_64bit", "arm64": "ubuntu_arm64"}[runtime.GOARCH]
	if flavour == "" {
		return fmt.Errorf("no Session Manager plugin build is available for linux/%s", runtime.GOARCH)
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

	debPath := filepath.Join(tmp, "session-manager-plugin.deb")
	url := fmt.Sprintf("https://s3.amazonaws.com/session-manager-downloads/plugin/latest/%s/session-manager-plugin.deb", flavour)
	if err := download(ctx, url, debPath); err != nil {
		return err
	}

	installDir := filepath.Join(home, ".local", "sessionmanagerplugin")
	if err := extractDeb(debPath, installDir, "usr/local/sessionmanagerplugin/"); err != nil {
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
