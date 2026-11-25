# Built-in Markdown Manual Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a 14-page beginner Markdown manual inside the OpenMD executable, reachable from the reading page's About tab and generated/reused locally without network access.

**Architecture:** Markdown chapters are embedded with `go:embed`. A new `internal/manual` package fingerprints the embedded files, materializes them to a content-addressed source directory, and renders them through the existing `render.BuildDirSite` pipeline. `app.openPage` ensures the manual exists before rendering the document site and passes its `file:///` URL into `render.SiteOptions`; the About tab renders a normal new-tab link.

**Tech Stack:** Go 1.26+, `go:embed`, existing goldmark/Chroma renderer, embedded reader HTML/CSS/JS, `file:///` pages.

**Spec:** `openspec/changes/builtin-manual/design.md` (with `proposal.md` and `specs/user-manual/spec.md`)

## Global Constraints

- Windows-only project; keep the existing `file:///` architecture and CSP.
- Manual prose is Simplified Chinese; OpenSpec structural headings stay English.
- Do not modify the installer, install/manage dialog, right-click menu, or add a `--manual` flag.
- Manual body must be original; external tutorials appear only as labeled links. Do not add entries to `THIRD_PARTY_NOTICES.md`.
- Manual pages must render with `ManualURL` empty so the manual's own About tab has no self-link.
- No remote images or network resources in the manual; examples use text, code fences, and tables only.
- Bump `internal/render/site.go` `cacheGeneration` whenever templates, CSS, or shared JS change.
- After each task: run `gofmt` on touched Go files, `go vet ./...`, `go test ./... -count=1`, then commit. Do not push.
- Manual generation failure must never block document opening; hide the About entry instead.
- Do not delete directory trees that may still be in use; pruning only removes non-current entries older than `pruneAge`.

## Review Focus

These are the inputs/failure modes most likely to bite a user; each has a test in the owning task.

1. **Two files opened at once on first run** — one process builds while the other waits; the second must reuse the finished manual or hide the entry, never corrupt it. Tested in Task 2 (`TestEnsureReusesExistingManualWhenLocked`, `TestEnsureReturnsErrorWhenLockedWithoutCurrent`).
2. **Process killed during chapter extraction** — a directory with the right fingerprint but missing chapters must not be reused. Tested in Task 1 (`TestExtractReplacesIncompleteDirectory`).
3. **Primary cache directory unwritable** — the document falls back to `%TEMP%\OpenMD`; the manual must be generated in the same fallback base. Tested in Task 4 (`TestOpenPageTriesManualForFallbackBase`).
4. **Cached document pages keep an old manual URL** — changing the manual fingerprint/URL must rebuild document pages. Tested in Task 3 (`TestBuildFileSiteRebuildsWhenManualURLChanges`).
5. **Content links, images, and attribution** — broken relative chapter links, remote images, or copied third-party prose must fail review/tests. Tested in Task 5.

---

### Task 1: File URL helper and manual extraction primitives

**Files:**
- Modify: `internal/render/url.go`
- Modify: `internal/render/url_test.go`
- Create: `internal/manual/manual.go`
- Create: `internal/manual/content/index.md`
- Create: `internal/manual/manual_test.go`

**Interfaces:**
- Produces: `render.FileURL(path string) string`
- Produces (package `manual`, unexported): `contentFS() (fs.FS, error)`, `fingerprint(fsys fs.FS) (string, error)`, `extract(fsys fs.FS, dst string) error`, `complete(dir string) bool`, `fileExists(path string) bool`
- Later tasks use `render.FileURL`, `contentFS`, `fingerprint`, `extract`, `complete`.

- [ ] **Step 1: Write the failing `FileURL` test**

Add to `internal/render/url_test.go`:

```go
func TestFileURLFormatsWindowsPath(t *testing.T) {
	if got := FileURL(`C:\docs\a b.md`); got != "file:///C:/docs/a%20b.md" {
		t.Fatalf("FileURL = %q", got)
	}
}
```

- [ ] **Step 2: Run it and verify it fails**

Run: `go test ./internal/render -run TestFileURLFormatsWindowsPath -count=1`
Expected: FAIL, `undefined: FileURL`.

- [ ] **Step 3: Implement `FileURL`**

In `internal/render/url.go`, add:

```go
// FileURL returns the file:/// URL for an absolute local path.
func FileURL(path string) string {
	u, _ := fileURL(path, "", "")
	return u
}
```

- [ ] **Step 4: Run the test and verify it passes**

Run: `go test ./internal/render -run TestFileURLFormatsWindowsPath -count=1`
Expected: PASS.

- [ ] **Step 5: Create the embeddable content placeholder**

Create `internal/manual/content/index.md`:

```markdown
# OpenMD 使用手册

手册正在编写中。
```

- [ ] **Step 6: Write failing fingerprint/extraction tests**

Create `internal/manual/manual_test.go`:

```go
package manual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
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
```

- [ ] **Step 7: Run tests and verify they fail**

Run: `go test ./internal/manual -count=1`
Expected: FAIL with undefined `fingerprint`, `extract`, or `complete`.

- [ ] **Step 8: Implement the manual package primitives**

Create `internal/manual/manual.go` with:

```go
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
```

- [ ] **Step 9: Run tests and verify they pass**

Run: `go test ./internal/manual ./internal/render -run 'TestFingerprint|TestExtract|TestFileURL' -count=1`
Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add internal/render/url.go internal/render/url_test.go internal/manual
git commit -m "feat(manual): add content fingerprint and extraction primitives"
```

---

### Task 2: Ensure the manual site exists and is reused

**Files:**
- Modify: `internal/manual/manual.go`
- Modify: `internal/manual/manual_test.go`

**Interfaces:**
- Consumes: `render.BuildDirSite`, `render.SiteOptions`, `render.FileURL`, `contentFS`, `fingerprint`, `extract`, `complete` (Task 1).
- Produces: `manual.Ensure(baseDir string) (Result, error)`, `manual.Result{URL, Fingerprint string}`; unexported `ensure(baseDir string, fsys fs.FS, now func() time.Time) (Result, error)`.
- Produces test knobs as package vars: `lockWait`, `lockStale`, `pruneAge`.

- [ ] **Step 1: Write failing `Ensure` tests**

Append to `internal/manual/manual_test.go`:

```go
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
```

Add `"time"` and `"path/filepath"` imports if not already present.

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./internal/manual -run TestEnsure -count=1`
Expected: FAIL, `undefined: ensure` / `readCurrent`.

- [ ] **Step 3: Implement `Ensure` and helpers**

In `internal/manual/manual.go`, add:

```go
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
	if !ok || cur.Fingerprint != fp || !fileExists(cur.InitialPage) {
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
```

Add imports `encoding/json`, `strings`, `time`, and `openmd/internal/render`.

Note: `fileExists(oldCache)` returns false for directories because it checks `!info.IsDir()`; the prune test should use `os.Stat`. Adjust the test in Step 1 to use `os.Stat(oldSrc)`/`os.Stat(oldCache)` instead of `fileExists`. Keep this in mind when writing the test.

- [ ] **Step 4: Run tests and fix compile or behavior issues**

Run: `go test ./internal/manual -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manual
git commit -m "feat(manual): generate and reuse the manual site"
```

---

### Task 3: Render the manual link into document pages

**Files:**
- Modify: `internal/render/site.go`
- Modify: `internal/render/docpage.go`
- Modify: `internal/render/page.go`
- Modify: `internal/render/web/docpage.html`
- Modify: `internal/render/web/page.html`
- Modify: `internal/render/web/reader.css`
- Modify: `internal/render/page_test.go`
- Modify: `internal/render/site_test.go`

**Interfaces:**
- Produces: `render.SiteOptions.ManualURL string`
- Produces: `render.DocPageData.ManualURL string`
- Produces: `render.Page.ManualURL string`
- Changes internal signatures: `manifestFor(version, knownHash, root, manualURL string, docs []*siteDoc)`, `renderDocs(..., hasFilesPanel bool, manualURL string)`, `renderOneDoc(..., hasFilesPanel bool, manualURL string)`.

- [ ] **Step 1: Write failing render/template tests**

Add to `internal/render/page_test.go`:

```go
func TestDocPageShowsManualLink(t *testing.T) {
	out, err := RenderDocPage(DocPageData{
		Title:     "A",
		ManualURL: "file:///C:/manual/doc-x.html",
		Doc:       Document{Title: "A", ID: "doc"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`href="file:///C:/manual/doc-x.html"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		"打开使用手册",
	} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("manual link missing %q", want)
		}
	}
}

func TestDocPageHidesManualLinkWithoutURL(t *testing.T) {
	out, err := RenderDocPage(DocPageData{Title: "A", Doc: Document{Title: "A", ID: "doc"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "打开使用手册") {
		t.Fatal("manual link must be hidden without a URL")
	}
}

func TestRenderPageShowsManualLink(t *testing.T) {
	out, err := RenderPage(&Page{
		Mode:      "file",
		Title:     "A",
		ManualURL: "file:///C:/manual/doc-x.html",
		Documents: []Document{{ID: "doc-0", Title: "A"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "打开使用手册") {
		t.Fatal("legacy page template must show the manual link too")
	}
}
```

- [ ] **Step 2: Run them and verify they fail**

Run: `go test ./internal/render -run 'TestDocPage.*ManualLink|TestRenderPageShowsManualLink' -count=1`
Expected: FAIL, `ManualURL` field or link missing.

- [ ] **Step 3: Add `ManualURL` to page data and templates**

- `DocPageData`: add `ManualURL string`.
- `pageData`: add `ManualURL string`.
- `Page`: add `ManualURL string`; in `RenderPage`, set `ManualURL: page.ManualURL`.
- In both `docpage.html` and `page.html`, inside `#settings-panel-about`, before the release block, add:

```html
{{if .ManualURL}}
<a class="manual-link panel-action" href="{{.ManualURL}}" target="_blank" rel="noopener noreferrer">打开使用手册</a>
{{end}}
```

- In `reader.css`, next to `.release-link`, add:

```css
.manual-link {
  margin: 10px 8px 0 0;
}
```

- [ ] **Step 4: Run template tests and verify they pass**

Run: `go test ./internal/render -run 'TestDocPage.*ManualLink|TestRenderPageShowsManualLink' -count=1`
Expected: PASS.

- [ ] **Step 5: Write the failing site cache test**

Add to `internal/render/site_test.go`:

```go
func TestBuildFileSiteRebuildsWhenManualURLChanges(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "note.md")
	if err := os.WriteFile(md, []byte("# Note\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cache := t.TempDir()
	opts := SiteOptions{CacheBase: cache, Version: "v1", AllowCache: true, ManualURL: "file:///C:/manual/a.html"}
	first, err := BuildFileSite(md, opts)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := os.ReadFile(first.InitialPage)
	if !strings.Contains(string(page), "file:///C:/manual/a.html") {
		t.Fatal("first manual URL missing")
	}
	opts.ManualURL = "file:///C:/manual/b.html"
	second, err := BuildFileSite(md, opts)
	if err != nil {
		t.Fatal(err)
	}
	page, _ = os.ReadFile(second.InitialPage)
	if !strings.Contains(string(page), "file:///C:/manual/b.html") || strings.Contains(string(page), "file:///C:/manual/a.html") {
		t.Fatal("cached page kept the old manual URL")
	}
	if second.Generated == 0 {
		t.Fatal("changing the manual URL must regenerate pages")
	}
}
```

- [ ] **Step 6: Run it and verify it fails**

Run: `go test ./internal/render -run TestBuildFileSiteRebuildsWhenManualURLChanges -count=1`
Expected: FAIL, `ManualURL` not applied.

- [ ] **Step 7: Thread `ManualURL` through the site renderer**

- `SiteOptions`: add `ManualURL string`.
- `siteManifest`: add `ManualURL string `json:"manualURL,omitempty"``.
- `manifestFor`: add a `manualURL string` parameter and set `ManualURL: manualURL`.
- `BuildDirSite` and `BuildFileSite`: include `manifest.ManualURL != opts.ManualURL` in `rebuildAll`; pass `opts.ManualURL` to `renderDocs` and to `manifestFor`.
- `renderDocs` and `renderOneDoc`: add a `manualURL string` parameter; `renderOneDoc` sets `ManualURL: manualURL` in `DocPageData`.
- Bump `cacheGeneration` to `"2026-11-25-builtin-manual"`.

- [ ] **Step 8: Run the full render test package**

Run: `go test ./internal/render -count=1`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add internal/render
git commit -m "feat(reader): expose the manual in the about panel"
```

---

### Task 4: Ensure the manual before opening documents

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

**Interfaces:**
- Consumes: `manual.Ensure`, `render.SiteOptions.ManualURL` (Tasks 2 and 3).
- Produces: `App.Manual func(baseDir string) (string, error)`; `manualURLFor(baseDir string) (string, error)`.

- [ ] **Step 1: Write failing app tests**

Add to `internal/app/run_test.go`:

```go
func TestOpenPageAddsManualLink(t *testing.T) {
	md := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(md, []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opener := &fakeOpener{}
	a := &App{
		Temp:   tempdir.New(t.TempDir()),
		Opener: opener,
		Manual: func(string) (string, error) { return "file:///C:/manual/doc-x.html", nil },
	}
	if err := a.Run(Invocation{Mode: ModeReadFile, Path: md}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(opener.path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "file:///C:/manual/doc-x.html") || !strings.Contains(string(data), "打开使用手册") {
		t.Fatal("manual link missing from opened page")
	}
}

func TestOpenPageHidesManualWhenEnsureFails(t *testing.T) {
	md := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(md, []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opener := &fakeOpener{}
	a := &App{
		Temp:   tempdir.New(t.TempDir()),
		Opener: opener,
		Manual: func(string) (string, error) { return "", errors.New("boom") },
	}
	if err := a.Run(Invocation{Mode: ModeReadFile, Path: md}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(opener.path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "打开使用手册") {
		t.Fatal("manual link must be hidden when Ensure fails")
	}
}

func TestOpenPageTriesManualForFallbackBase(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TEMP", tmp)
	t.Setenv("TMP", tmp)

	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	md := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(md, []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opener := &fakeOpener{}
	var bases []string
	a := &App{
		Temp:   tempdir.New(blocked),
		Opener: opener,
		Manual: func(base string) (string, error) {
			bases = append(bases, base)
			return "file:///C:/manual/doc-x.html", nil
		},
	}
	if err := a.Run(Invocation{Mode: ModeReadFile, Path: md}); err != nil {
		t.Fatal(err)
	}
	fallback := filepath.Join(tmp, "OpenMD")
	found := false
	for _, base := range bases {
		if base == fallback {
			found = true
		}
	}
	if !found {
		t.Fatalf("manual was not attempted in fallback base: %#v", bases)
	}
	data, _ := os.ReadFile(opener.path)
	if !strings.Contains(string(data), "打开使用手册") {
		t.Fatal("fallback page lost the manual link")
	}
}
```

Add `"errors"` to the test imports.

- [ ] **Step 2: Run them and verify they fail**

Run: `go test ./internal/app -run TestOpenPage -count=1`
Expected: FAIL, `App.Manual` undefined or link missing.

- [ ] **Step 3: Implement the dependency and wiring**

- In `internal/app/app.go`, add `Manual func(baseDir string) (string, error)` to `App`.
- In `internal/app/run.go`:

```go
func (a *App) openPage(path string, isDir bool) error {
	...
	build := func(base string) (*render.Site, error) {
		manualURL := ""
		if a.Manual != nil {
			if url, err := a.Manual(base); err == nil {
				manualURL = url
			}
		}
		opts := render.SiteOptions{
			CacheBase:  filepath.Join(base, "rendered"),
			Version:    buildinfo.VersionLabel(),
			AllowCache: buildinfo.IsReleaseBuild(),
			Confirm:    a.confirmLoad,
			ManualURL:  manualURL,
		}
		...
	}
	...
}

func manualURLFor(baseDir string) (string, error) {
	result, err := manual.Ensure(baseDir)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
```

- In `Main`, set `Manual: manualURLFor` on the `App`.
- Import `openmd/internal/manual` and `errors` as needed.

- [ ] **Step 4: Run the app tests and verify they pass**

Run: `go test ./internal/app -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/app
git commit -m "feat(app): prepare the manual before opening documents"
```

---

### Task 5: Write the manual chapters

**Files:**
- Replace: `internal/manual/content/index.md`
- Create: `internal/manual/content/01-what-is-markdown.md`
- Create: `internal/manual/content/02-headings-paragraphs.md`
- Create: `internal/manual/content/03-emphasis-inline-code.md`
- Create: `internal/manual/content/04-lists.md`
- Create: `internal/manual/content/05-links-images.md`
- Create: `internal/manual/content/06-blockquotes.md`
- Create: `internal/manual/content/07-code-blocks.md`
- Create: `internal/manual/content/08-tables.md`
- Create: `internal/manual/content/09-rules-and-html.md`
- Create: `internal/manual/content/10-organize-documents.md`
- Create: `internal/manual/content/11-recipes.md`
- Create: `internal/manual/content/12-cheatsheet.md`
- Create: `internal/manual/content/13-faq.md`
- Create: `internal/manual/content_test.go`

**Interfaces:**
- Consumes: `contentFS()` (Task 1).
- Produces: the embedded content used by `Ensure`.

**Content contract (every chapter file, 01–13):**
- Exactly one H1, starting with `# `.
- Contains these five H2 headings in order: `## 你会学到`, `## 基本写法`, `## 在 OpenMD 里怎么看`, `## 常见错误`, `## 延伸阅读`.
- At least one fenced code block showing Markdown source.
- The `延伸阅读` section has at least one `https://` link.
- No remote images and no `<img>` tags.

**Chapter topics:**

| File | Required topics |
| --- | --- |
| `index.md` | What the manual is, who it is for, how to read with 阅读/源码/双视图, a linked table of contents for all 13 chapters, and a note that the prose is original with external links as further reading. |
| `01-what-is-markdown.md` | Plain text plus symbols; why local Markdown; the three OpenMD views and how to switch; a first title + paragraph example. |
| `02-headings-paragraphs.md` | `#` to `######`; one H1; blank-line paragraph rules; hard line breaks; how headings drive the outline and document title. |
| `03-emphasis-inline-code.md` | Bold, italic, both, strikethrough, inline code, escaping with `\`; spaces inside markers. |
| `04-lists.md` | Unordered markers, ordered lists, nesting indentation, task lists, tight vs loose lists. |
| `05-links-images.md` | Inline links, titles, relative links, cross-document links, heading anchors, images and alt text, paths with spaces. |
| `06-blockquotes.md` | Single and nested quotes, multi-paragraph quotes, lists and code inside quotes. |
| `07-code-blocks.md` | Inline code, fenced blocks, language tags, highlighting, long lines, backticks inside code. |
| `08-tables.md` | Pipe tables, alignment separators, escaping pipes, wide-table scrolling and print behavior. |
| `09-rules-and-html.md` | Horizontal rules, `<details>`/`<summary>`, when HTML is acceptable, when to prefer Markdown. |
| `10-organize-documents.md` | README/index priority in directory mode, directory layout, front matter stripping, relative cross-references, moving files. |
| `11-recipes.md` | Five runnable templates: project README, study notes, API reference, meeting notes, long-form printing. |
| `12-cheatsheet.md` | One compact table covering headings, emphasis, lists, links, images, code, quotes, tables, rules, task lists, details. |
| `13-faq.md` | At least: why a single newline does not break lines; why a table does not render; why code has no highlighting; Chinese punctuation in links; Windows paths; broken images; front matter visibility; heading-level jumps in the outline. |

Use half-width Markdown punctuation in examples. Link to chapters with relative `.md` paths, for example `[下一章](02-headings-paragraphs.md)`.

- [ ] **Step 1: Write failing content tests**

Create `internal/manual/content_test.go`:

```go
package manual

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

var chapterFiles = []string{
	"01-what-is-markdown.md",
	"02-headings-paragraphs.md",
	"03-emphasis-inline-code.md",
	"04-lists.md",
	"05-links-images.md",
	"06-blockquotes.md",
	"07-code-blocks.md",
	"08-tables.md",
	"09-rules-and-html.md",
	"10-organize-documents.md",
	"11-recipes.md",
	"12-cheatsheet.md",
	"13-faq.md",
}

func TestManualContentStructure(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		text := string(data)
		if len(regexp.MustCompile(`(?m)^# `).FindAllString(text, -1)) != 1 {
			t.Fatalf("%s must have exactly one H1", name)
		}
		if name == "index.md" {
			continue
		}
		for _, heading := range []string{
			"## 你会学到",
			"## 基本写法",
			"## 在 OpenMD 里怎么看",
			"## 常见错误",
			"## 延伸阅读",
		} {
			if !strings.Contains(text, heading) {
				t.Fatalf("%s missing %q", name, heading)
			}
		}
		if !strings.Contains(text, "```") {
			t.Fatalf("%s must contain a fenced example", name)
		}
		if !strings.Contains(text, "https://") {
			t.Fatalf("%s needs a further-reading link", name)
		}
	}
}

func TestManualIndexLinksEveryChapter(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(fsys, "index.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, name := range chapterFiles {
		if !strings.Contains(text, "]("+name+")") {
			t.Fatalf("index.md does not link to %s", name)
		}
	}
}

var markdownLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

func TestManualRelativeLinksResolve(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range markdownLink.FindAllStringSubmatch(string(data), -1) {
			target := match[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			if idx := strings.Index(target, "#"); idx >= 0 {
				target = target[:idx]
			}
			if !strings.HasSuffix(target, ".md") {
				continue
			}
			resolved := path.Clean(path.Join(path.Dir(name), target))
			if _, err := fs.Stat(fsys, resolved); err != nil {
				t.Fatalf("%s links to missing %s", name, resolved)
			}
		}
	}
}

func TestManualHasNoRemoteImages(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, "<img") || regexp.MustCompile(`!\[[^\]]*\]\(https?://`).MatchString(text) {
			t.Fatalf("%s must not reference remote images", name)
		}
	}
}
```

- [ ] **Step 2: Run content tests and verify they fail**

Run: `go test ./internal/manual -run TestManual -count=1`
Expected: FAIL, missing chapter files and sections.

- [ ] **Step 3: Write the chapters**

Write all 14 files following the content contract and topics table. Keep each chapter beginner-sized (roughly 400–800 Chinese characters plus examples). Use only the listed further-reading sites:
- `https://www.markdownguide.org/`
- `https://commonmark.org/`
- `https://docs.github.com/zh/get-started/writing-on-github`

The index must link every chapter, and every chapter should end with `## 延伸阅读`.

- [ ] **Step 4: Run content tests and verify they pass**

Run: `go test ./internal/manual -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manual/content internal/manual/content_test.go
git commit -m "docs(manual): add the beginner Markdown chapters"
```

---

### Task 6: Documentation, end-to-end verification, and OpenSpec record

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `openspec/changes/builtin-manual/tasks.md`

**Interfaces:**
- No new code interfaces. This task proves the assembled feature and records the result.

- [ ] **Step 1: Add the README capability bullet**

In `README.md` under `## 更多能力`, add:

```markdown
- 内置 Markdown 使用手册：在「信息与设置 → 关于」里打开，从零学习常用写法。
```

- [ ] **Step 2: Document the manual cache layout**

In `AGENTS.md` under `## 临时文件`, add a short block after the rendered cache example:

```text
%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\manual\
  current.json
  src-<content-hash>\
  cache\<root-hash>\
```

Add one sentence: `current.json` remembers the active content fingerprint; old `src-*` and cache roots older than 24 hours are pruned.

- [ ] **Step 3: Run the full verification**

Run:

```bash
gofmt -w internal/manual internal/app internal/render
go vet ./...
go test ./... -count=1
```

Expected: all packages pass.

- [ ] **Step 4: Verify the assembled feature in a browser**

Render a temporary document with the real manual:

1. Create a temporary Markdown file and a temporary cache base.
2. Use a tiny Go program (or `go test` helper) to call `manual.Ensure(cacheBase)` and `render.BuildFileSite(md, render.SiteOptions{CacheBase: cacheBase, Version: "verify", AllowCache: true, ManualURL: result.URL})`.
3. Open the generated page in a browser.
4. Open `信息与设置 → 关于`, confirm `打开使用手册` is present.
5. Click it; confirm a new tab opens the manual index; click a chapter link and the sidebar chapter list; switch to 源码 and 双视图.
6. Temporarily pass an invalid `ManualURL` (or empty) and confirm the entry is hidden.

Record the actual observed results in `openspec/changes/builtin-manual/tasks.md`.

- [ ] **Step 5: Update OpenSpec tasks and verification record**

Check off every task in `openspec/changes/builtin-manual/tasks.md` and append a `## 验证记录` section with the commands run and the browser observations.

- [ ] **Step 6: Commit**

```bash
git add README.md AGENTS.md openspec/changes/builtin-manual
git commit -m "docs: document the built-in manual"
```

---

## Self-Review

- **Spec coverage:** entry in About tab → Task 3; offline embedded manual → Tasks 1, 2, 5; generation does not block reading → Task 4; reuse/fingerprint → Task 2; failure hides entry → Task 4; original content with attributed links → Task 5; concurrent opens → Task 2; fallback base → Task 4; cache invalidation → Task 3.
- **Type consistency:** `manual.Ensure`/`Result`, `render.FileURL`, `SiteOptions.ManualURL`, `DocPageData.ManualURL`, `Page.ManualURL`, `App.Manual` are used consistently across tasks.
- **Review Focus:** all five failure modes map to named tests in Tasks 1–5.
