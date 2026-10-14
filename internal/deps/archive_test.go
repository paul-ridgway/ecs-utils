package deps

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// buildDeb assembles a minimal .deb (ar archive with a gzipped data tarball).
func buildDeb(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var data bytes.Buffer
	gz := gzip.NewWriter(&data)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	var ar bytes.Buffer
	ar.WriteString("!<arch>\n")
	addMember := func(name string, body []byte) {
		fmt.Fprintf(&ar, "%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "100644", len(body))
		ar.Write(body)
		if len(body)%2 == 1 {
			ar.WriteByte('\n')
		}
	}
	addMember("debian-binary", []byte("2.0\n"))
	addMember("control.tar.gz", []byte("x"))
	addMember("data.tar.gz/", data.Bytes())
	return ar.Bytes()
}

func TestExtractDeb(t *testing.T) {
	deb := buildDeb(t, map[string]string{
		"./usr/local/sessionmanagerplugin/bin/session-manager-plugin": "#!/bin/sh\necho hi\n",
		"./usr/share/doc/readme": "ignored",
	})
	dir := t.TempDir()
	debPath := filepath.Join(dir, "test.deb")
	if err := os.WriteFile(debPath, deb, 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "out")
	if err := extractDeb(debPath, dest, "usr/local/sessionmanagerplugin/"); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dest, "bin", "session-manager-plugin")
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("expected executable bit on %s, got %v", bin, info.Mode())
	}
	if _, err := os.Stat(filepath.Join(dest, "usr")); !os.IsNotExist(err) {
		t.Error("files outside the prefix should not be extracted")
	}
}

func TestFindArMemberMissing(t *testing.T) {
	deb := buildDeb(t, map[string]string{"a": "b"})
	if _, _, err := findArMember(bufio.NewReader(bytes.NewReader(deb)), "nope"); err == nil {
		t.Error("expected an error for a missing member")
	}
}

func TestSafeJoinRejectsEscape(t *testing.T) {
	if _, err := safeJoin(t.TempDir(), "../../etc/passwd"); err == nil {
		t.Error("expected path escape to be rejected")
	}
}
