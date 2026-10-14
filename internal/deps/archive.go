package deps

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// extractZip unpacks the entries of zipPath that live under prefix into dest,
// stripping the prefix. File modes and symlinks are preserved.
func extractZip(zipPath, dest, prefix string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rel, ok := strings.CutPrefix(f.Name, prefix)
		if !ok || rel == "" {
			continue
		}
		target, err := safeJoin(dest, rel)
		if err != nil {
			return err
		}
		mode := f.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case mode&os.ModeSymlink != 0:
			rc, err := f.Open()
			if err != nil {
				return err
			}
			linkTarget, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}
			if err := replaceSymlink(string(linkTarget), target); err != nil {
				return err
			}
		default:
			rc, err := f.Open()
			if err != nil {
				return err
			}
			err = writeFile(target, rc, mode.Perm())
			rc.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// extractDeb unpacks the files under prefix from a Debian package's data
// archive into dest, stripping the prefix. No dpkg is required.
func extractDeb(debPath, dest, prefix string) error {
	f, err := os.Open(debPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, name, err := findArMember(bufio.NewReader(f), "data.tar")
	if err != nil {
		return err
	}
	var tr io.Reader = data
	switch {
	case strings.HasSuffix(name, ".gz"):
		gz, err := gzip.NewReader(data)
		if err != nil {
			return err
		}
		defer gz.Close()
		tr = gz
	case strings.HasSuffix(name, ".tar"):
	default:
		return fmt.Errorf("unsupported data archive %q in %s", name, debPath)
	}
	return extractTar(tr, dest, prefix)
}

// findArMember walks a GNU/BSD "ar" archive and returns a reader over the
// first member whose name starts with namePrefix.
func findArMember(r *bufio.Reader, namePrefix string) (io.Reader, string, error) {
	magic := make([]byte, 8)
	if _, err := io.ReadFull(r, magic); err != nil || string(magic) != "!<arch>\n" {
		return nil, "", errors.New("not an ar archive")
	}
	header := make([]byte, 60)
	for {
		if _, err := io.ReadFull(r, header); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, "", fmt.Errorf("member %s* not found in archive", namePrefix)
			}
			return nil, "", err
		}
		name := strings.TrimSuffix(strings.TrimSpace(string(header[0:16])), "/")
		size, err := strconv.ParseInt(strings.TrimSpace(string(header[48:58])), 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("bad ar header: %w", err)
		}
		if strings.HasPrefix(name, namePrefix) {
			return io.LimitReader(r, size), name, nil
		}
		skip := size + size%2 // members are padded to an even length
		if _, err := io.CopyN(io.Discard, r, skip); err != nil {
			return nil, "", err
		}
	}
}

// extractTar unpacks the entries under prefix into dest, stripping the prefix.
func extractTar(r io.Reader, dest, prefix string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(hdr.Name, "./")
		rel, ok := strings.CutPrefix(name, prefix)
		if !ok || rel == "" {
			continue
		}
		target, err := safeJoin(dest, rel)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeFile(target, tr, os.FileMode(hdr.Mode).Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := replaceSymlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
}

func writeFile(target string, r io.Reader, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if perm == 0 {
		perm = 0o644
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Chmod(target, perm)
}

// safeJoin joins rel onto dest, refusing paths that escape dest.
func safeJoin(dest, rel string) (string, error) {
	target := filepath.Join(dest, filepath.FromSlash(rel))
	if target != dest && !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes destination", rel)
	}
	return target, nil
}
