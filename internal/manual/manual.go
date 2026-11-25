package manual

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"openmd/internal/render"
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

type Result struct {
	URL         string
	Fingerprint string
}

type currentManual struct {
	Fingerprint string `json:"fingerprint"`
	InitialPage string `json:"initialPage"`
}

var (
	lockWait  = 2 * time.Second
	lockStale = 5 * time.Minute
	pruneAge  = 24 * time.Hour
)

// Ensure materializes the embedded manual under baseDir and returns its entry URL.
func Ensure(baseDir string) (Result, error) {
	fsys, err := contentFS()
	if err != nil {
		return Result{}, err
	}
	return ensure(baseDir, fsys, time.Now)
}

func ensure(baseDir string, fsys fs.FS, now func() time.Time) (Result, error) {
	fp, err := fingerprint(fsys)
	if err != nil {
		return Result{}, err
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return Result{}, err
	}
	manualDir := filepath.Join(baseDir, "manual")
	if err := os.MkdirAll(manualDir, 0755); err != nil {
		return Result{}, err
	}
	if res, ok := currentResult(manualDir, fp); ok {
		return res, nil
	}

	release, err := acquireLock(filepath.Join(manualDir, ".lock"), now)
	if err != nil {
		if res, ok := currentResult(manualDir, fp); ok {
			return res, nil
		}
		return Result{}, err
	}
	defer release()

	if res, ok := currentResult(manualDir, fp); ok {
		return res, nil
	}

	srcDir := filepath.Join(manualDir, "src-"+fp)
	if err := extract(fsys, srcDir); err != nil {
		return Result{}, err
	}
	site, err := render.BuildDirSite(srcDir, render.SiteOptions{
		CacheBase:  filepath.Join(manualDir, "cache"),
		Version:    "manual",
		AllowCache: true,
	})
	if err != nil {
		return Result{}, err
	}
	if err := writeCurrent(manualDir, currentManual{Fingerprint: fp, InitialPage: site.InitialPage}); err != nil {
		return Result{}, err
	}
	prune(manualDir, srcDir, filepath.Dir(site.InitialPage), now())
	return Result{URL: render.FileURL(site.InitialPage), Fingerprint: fp}, nil
}

func currentResult(manualDir, fp string) (Result, bool) {
	cur, ok := readCurrent(manualDir)
	if !ok || cur.Fingerprint != fp || !filepath.IsAbs(cur.InitialPage) || !fileExists(cur.InitialPage) {
		return Result{}, false
	}
	return Result{URL: render.FileURL(cur.InitialPage), Fingerprint: fp}, true
}

func readCurrent(manualDir string) (currentManual, bool) {
	data, err := os.ReadFile(filepath.Join(manualDir, "current.json"))
	if err != nil {
		return currentManual{}, false
	}
	var cur currentManual
	if err := json.Unmarshal(data, &cur); err != nil || cur.Fingerprint == "" || cur.InitialPage == "" {
		return currentManual{}, false
	}
	return cur, true
}

func writeCurrent(manualDir string, cur currentManual) error {
	data, err := json.Marshal(cur)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(manualDir, ".current-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	return os.Rename(name, filepath.Join(manualDir, "current.json"))
}

func acquireLock(path string, now func() time.Time) (func(), error) {
	deadline := now().Add(lockWait)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			f.Close()
			return func() { _ = os.Remove(path) }, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		if info, statErr := os.Stat(path); statErr == nil && now().Sub(info.ModTime()) > lockStale {
			_ = os.Remove(path)
			continue
		}
		if now().After(deadline) {
			return nil, fmt.Errorf("manual generation is already in progress")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func prune(manualDir, keepSrc, keepCache string, now time.Time) {
	entries, err := os.ReadDir(manualDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(manualDir, name)
		switch {
		case strings.HasPrefix(name, "src-") && path != keepSrc:
			removeIfOld(path, now)
		case strings.HasPrefix(name, ".staging-"):
			removeIfOld(path, now)
		case name == "cache":
			roots, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, root := range roots {
				rootPath := filepath.Join(path, root.Name())
				if rootPath != keepCache {
					removeIfOld(rootPath, now)
				}
			}
		}
	}
}

func removeIfOld(path string, now time.Time) {
	info, err := os.Stat(path)
	if err == nil && now.Sub(info.ModTime()) > pruneAge {
		_ = os.RemoveAll(path)
	}
}
