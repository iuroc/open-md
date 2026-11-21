package tempdir

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Manager struct {
	BaseDir string
	Now     func() time.Time
	MaxAge  time.Duration
	MaxRuns int
}

func New(baseDir string) *Manager {
	return &Manager{
		BaseDir: baseDir,
		Now:     time.Now,
		MaxAge:  7 * 24 * time.Hour,
		MaxRuns: 50,
	}
}

func (m *Manager) RunDir() (string, error) {
	if err := os.MkdirAll(m.BaseDir, 0755); err != nil {
		return "", err
	}
	now := time.Now
	if m.Now != nil {
		now = m.Now
	}
	return os.MkdirTemp(m.BaseDir, "run-"+now().Format("20060102-150405")+"-*")
}

// RenderRoot returns the stable cache directory for a scan root key.
func (m *Manager) RenderRoot(key string) (string, error) {
	path := filepath.Join(m.BaseDir, "rendered", key)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}

func (m *Manager) Cleanup() error {
	entries, err := os.ReadDir(m.BaseDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	type runEntry struct {
		path    string
		modTime time.Time
	}
	runs := make([]runEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "run-") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		runs = append(runs, runEntry{
			path:    filepath.Join(m.BaseDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].modTime.After(runs[j].modTime)
	})

	maxAge := m.MaxAge
	if maxAge <= 0 {
		maxAge = 7 * 24 * time.Hour
	}
	maxRuns := m.MaxRuns
	if maxRuns <= 0 {
		maxRuns = 50
	}
	now := time.Now
	if m.Now != nil {
		now = m.Now
	}
	cutoff := now().Add(-maxAge)
	for i, run := range runs {
		if i >= maxRuns || run.modTime.Before(cutoff) {
			_ = os.RemoveAll(run.path)
		}
	}
	return cleanRendered(filepath.Join(m.BaseDir, "rendered"), cutoff, maxRuns)
}

func cleanRendered(root string, cutoff time.Time, maxRoots int) error {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	type rootEntry struct {
		path    string
		modTime time.Time
	}
	roots := make([]rootEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		roots = append(roots, rootEntry{
			path:    filepath.Join(root, entry.Name()),
			modTime: info.ModTime(),
		})
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].modTime.After(roots[j].modTime)
	})
	for i, entry := range roots {
		if i >= maxRoots || entry.modTime.Before(cutoff) {
			_ = os.RemoveAll(entry.path)
		}
	}
	return nil
}
