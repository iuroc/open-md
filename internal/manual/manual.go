package manual

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

//go:embed content/*.md
var embeddedContent embed.FS

func contentFS() (fs.FS, error) {
	return fs.Sub(embeddedContent, "content")
}

func fingerprint(fsys fs.FS) (string, error) {
	names, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", name, len(data))
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

func complete(dir string) bool {
	return fileExists(filepath.Join(dir, ".complete"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func extract(fsys fs.FS, dst string) error {
	if complete(dst) {
		return nil
	}
	staging, err := os.MkdirTemp(filepath.Dir(dst), ".staging-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)

	names, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(staging, name), data, 0644); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(staging, ".complete"), []byte("ok\n"), 0644); err != nil {
		return err
	}
	if err := os.Rename(staging, dst); err == nil {
		return nil
	}
	if complete(dst) {
		return nil
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return os.Rename(staging, dst)
}
