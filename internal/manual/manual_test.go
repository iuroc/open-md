package manual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestFingerprintChangesWithContent(t *testing.T) {
	a := fstest.MapFS{"index.md": {Data: []byte("# A\n")}}
	b := fstest.MapFS{"index.md": {Data: []byte("# B\n")}}
	ha, err := fingerprint(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := fingerprint(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha == "" || ha == hb {
		t.Fatalf("fingerprints must differ: %q %q", ha, hb)
	}
}

func TestExtractWritesCompleteMarker(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "src-abc")
	fsys := fstest.MapFS{"index.md": {Data: []byte("# 测试手册\n")}}
	if err := extract(fsys, dst); err != nil {
		t.Fatal(err)
	}
	if !complete(dst) {
		t.Fatal("missing .complete marker")
	}
	data, err := os.ReadFile(filepath.Join(dst, "index.md"))
	if err != nil || !strings.Contains(string(data), "测试手册") {
		t.Fatalf("chapter not written: %v %q", err, data)
	}
}

func TestExtractReplacesIncompleteDirectory(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "src-abc")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "index.md"), []byte("# old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fsys := fstest.MapFS{"index.md": {Data: []byte("# new\n")}}
	if err := extract(fsys, dst); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dst, "index.md"))
	if !strings.Contains(string(data), "new") {
		t.Fatalf("incomplete directory was reused: %q", data)
	}
	if !complete(dst) {
		t.Fatal("missing .complete marker")
	}
}

func testContent() fstest.MapFS {
	return fstest.MapFS{
		"index.md": {Data: []byte("# 测试手册\n\n[第二章](02.md)\n")},
		"02.md":    {Data: []byte("# 第二章\n")},
	}
}

func TestEnsureBuildsManualSite(t *testing.T) {
	base := t.TempDir()
	res, err := ensure(base, testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if res.Fingerprint == "" || !strings.HasPrefix(res.URL, "file:///") {
		t.Fatalf("bad result: %#v", res)
	}
	cur, ok := readCurrent(filepath.Join(base, "manual"))
	if !ok || cur.Fingerprint != res.Fingerprint || !fileExists(cur.InitialPage) {
		t.Fatalf("current marker not usable: %#v", cur)
	}
	data, _ := os.ReadFile(cur.InitialPage)
	if !strings.Contains(string(data), "测试手册") {
		t.Fatalf("rendered manual missing title: %q", data)
	}
}

func TestEnsureReturnsAbsoluteURLForRelativeBase(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	res, err := ensure(".", testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.URL, "file:///") {
		t.Fatalf("manual URL must be absolute: %q", res.URL)
	}
	cur, ok := readCurrent(filepath.Join(dir, "manual"))
	if !ok || !filepath.IsAbs(cur.InitialPage) {
		t.Fatalf("initial page must be absolute: %#v", cur)
	}
}

func TestEnsureHealsRelativeCurrentMarker(t *testing.T) {
	base, err := os.MkdirTemp(".", ".manual-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)
	base, err = filepath.Abs(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ensure(base, testContent(), time.Now); err != nil {
		t.Fatal(err)
	}
	manualDir := filepath.Join(base, "manual")
	cur, ok := readCurrent(manualDir)
	if !ok {
		t.Fatal("current marker missing")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, cur.InitialPage)
	if err != nil {
		t.Fatal(err)
	}
	cur.InitialPage = rel
	if err := writeCurrent(manualDir, cur); err != nil {
		t.Fatal(err)
	}
	res, err := ensure(base, testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.URL, "file:///.") {
		t.Fatalf("manual URL still relative: %q", res.URL)
	}
	cur, ok = readCurrent(manualDir)
	if !ok || !filepath.IsAbs(cur.InitialPage) {
		t.Fatalf("relative marker was not healed: %#v", cur)
	}
}

func TestEnsureReusesCurrentManual(t *testing.T) {
	base := t.TempDir()
	first, err := ensure(base, testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ensure(base, testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if second.URL != first.URL || second.Fingerprint != first.Fingerprint {
		t.Fatalf("manual was not reused: %#v vs %#v", first, second)
	}
}

func TestEnsureReusesExistingManualWhenLocked(t *testing.T) {
	base := t.TempDir()
	first, err := ensure(base, testContent(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := os.OpenFile(filepath.Join(base, "manual", ".lock"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
	second, err := ensure(base, testContent(), time.Now)
	if err != nil || second.URL != first.URL {
		t.Fatalf("valid manual must be reused while locked: %#v %v", second, err)
	}
}

func TestEnsureReturnsErrorWhenLockedWithoutCurrent(t *testing.T) {
	old := lockWait
	lockWait = 20 * time.Millisecond
	defer func() { lockWait = old }()

	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "manual"), 0755); err != nil {
		t.Fatal(err)
	}
	lock, err := os.OpenFile(filepath.Join(base, "manual", ".lock"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
	if _, err := ensure(base, testContent(), time.Now); err == nil {
		t.Fatal("expected lock timeout error")
	}
}

func TestEnsureRemovesStaleLock(t *testing.T) {
	old := lockStale
	lockStale = time.Millisecond
	defer func() { lockStale = old }()

	base := t.TempDir()
	manualDir := filepath.Join(base, "manual")
	if err := os.MkdirAll(manualDir, 0755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(manualDir, ".lock")
	if err := os.WriteFile(lockPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lockPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if _, err := ensure(base, testContent(), time.Now); err != nil {
		t.Fatal(err)
	}
}

func TestEnsurePrunesOldVersions(t *testing.T) {
	old := pruneAge
	pruneAge = 0
	defer func() { pruneAge = old }()

	base := t.TempDir()
	if _, err := ensure(base, testContent(), time.Now); err != nil {
		t.Fatal(err)
	}
	manualDir := filepath.Join(base, "manual")
	oldSrc := filepath.Join(manualDir, "src-old")
	oldCache := filepath.Join(manualDir, "cache", "old-root")
	for _, dir := range []string{oldSrc, oldCache} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-time.Hour)
	for _, dir := range []string{oldSrc, oldCache} {
		if err := os.Chtimes(dir, past, past); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(manualDir, "current.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := ensure(base, testContent(), time.Now); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{oldSrc, oldCache} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("%s was not pruned", dir)
		}
	}
}
