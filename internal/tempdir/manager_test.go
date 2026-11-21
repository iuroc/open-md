package tempdir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunDirCreatesUniqueDirectories(t *testing.T) {
	m := New(t.TempDir())
	a, err := m.RunDir()
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.RunDir()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("run dirs are identical")
	}
	for _, p := range []string{a, b} {
		if filepath.Dir(p) != m.BaseDir {
			t.Fatalf("%q not under %q", p, m.BaseDir)
		}
		if !strings.HasPrefix(filepath.Base(p), "run-") {
			t.Fatalf("bad name %q", p)
		}
	}
}

func TestCleanupRemovesOldAndKeepsNew(t *testing.T) {
	base := t.TempDir()
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	m := New(base)
	m.Now = func() time.Time { return now }
	m.MaxAge = 7 * 24 * time.Hour

	old := filepath.Join(base, "run-old")
	fresh := filepath.Join(base, "run-fresh")
	if err := os.Mkdir(old, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(fresh, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(old, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(fresh, now, now); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("old run dir not removed")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatal("fresh run dir removed")
	}
}

func TestCleanupKeepsNewestFifty(t *testing.T) {
	base := t.TempDir()
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	m := New(base)
	m.Now = func() time.Time { return now }

	for i := 0; i < 55; i++ {
		name := filepath.Join(base, fmt.Sprintf("run-%02d", i))
		if err := os.Mkdir(name, 0755); err != nil {
			t.Fatal(err)
		}
		mod := now.Add(-time.Duration(i) * time.Minute)
		if err := os.Chtimes(name, mod, mod); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 50 {
		t.Fatalf("kept %d dirs, want 50", len(entries))
	}
}

func TestRenderRootIsStable(t *testing.T) {
	m := New(t.TempDir())
	a, err := m.RenderRoot("docs-abc123")
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.RenderRoot("docs-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("render roots differ: %q and %q", a, b)
	}
	if info, err := os.Stat(a); err != nil || !info.IsDir() {
		t.Fatalf("render root not created: %v", err)
	}
}

func TestCleanupRemovesOldRenderedRoots(t *testing.T) {
	base := t.TempDir()
	now := time.Now()
	m := New(base)
	m.Now = func() time.Time { return now }
	old, err := m.RenderRoot("old-root")
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := m.RenderRoot("fresh-root")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(old, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(fresh, now, now); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("old rendered root not removed")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatal("fresh rendered root removed")
	}
}

func TestCleanupIgnoresUnrelatedEntries(t *testing.T) {
	base := t.TempDir()
	m := New(base)
	if err := os.WriteFile(filepath.Join(base, "notes.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(base, "other"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "notes.txt")); err != nil {
		t.Fatal("unrelated file removed")
	}
	if _, err := os.Stat(filepath.Join(base, "other")); err != nil {
		t.Fatal("unrelated directory removed")
	}
}
