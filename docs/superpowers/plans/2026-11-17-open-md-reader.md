# OpenMD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single Windows exe that self-installs as a Markdown “open with” handler, renders Markdown to a temporary HTML file, and opens it in the default browser, including a folder reading mode with a switchable document list and TOC.

**Architecture:** Go renders Markdown with goldmark and Chroma, embeds CSS/JS into one generated HTML file, and launches the browser through `rundll32`. The exe copies itself to `%LOCALAPPDATA%\Programs\OpenMD`, registers per-user context-menu and “Open with” entries, and writes disposable run directories under its own install directory.

**Tech Stack:** Go 1.26+, goldmark, goldmark-highlighting/Chroma, `golang.org/x/sys/windows`, native Win32 MessageBox, embedded HTML/CSS/JS.

**Spec:** `docs/superpowers/specs/2026-11-17-open-md-reader-design.md`

**Execution note:** The user requested Native execution, no independent design review, tests before implementation, and manual acceptance at the end. Do not create git commits or push unless the user explicitly asks.

## Global Constraints

- Target Windows 10/11 amd64; no Linux/macOS support in this plan.
- Final artifact is one `OpenMD.exe`, built with `-H=windowsgui`; all assets are embedded.
- Registry writes are limited to `HKCU`; never require administrator rights.
- Never change the default `.md` or `.markdown` association.
- No HTTP server, no background service, no browser-based editing, no base64 image embedding in this version.
- Directory mode scans only the selected directory’s current level, never recursively.
- Temporary HTML goes only to `%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\...`; fallback is `%TEMP%\OpenMD\...`; never the Markdown file’s directory.
- Directory limits: 500 documents and 25 MB of raw Markdown.
- Raw HTML inside Markdown is not rendered.
- Chrome/Edge are the primary browsers; Firefox’s `file://` restriction is documented as a limitation.
- All user-facing copy is Chinese.
- Do not initialize git or commit unless the user asks.

## Review Focus

1. **Paths with spaces, Chinese characters, `#`, `%`, and UNC paths** — a correct-looking page with broken images is the most likely user-visible failure. Task 3 pins URL encoding; Task 10’s manual checklist confirms the browser behavior.
2. **Duplicate heading IDs across documents** — directory mode embeds multiple documents in one DOM; unprefixed IDs make TOC links jump to the wrong document. Task 5 pins per-document ID prefixes.
3. **Registry preservation** — installing or uninstalling must not remove another app’s `.md` association or overwrite the default handler. Task 8 tests against a temporary HKCU key.
4. **Unreadable or oversized directories** — one bad file must not break the whole page. Task 5 pins per-file placeholders and scan limits.
5. **Browser launch failure** — no default browser or a blocked handler should produce a clear error, not a silent no-op. Task 7 pins the `rundll32` → `explorer` fallback and error return.

---

### Task 1: Module bootstrap and argument parsing

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `internal/app/parse.go`
- Test: `internal/app/parse_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  package app

  type Mode int
  const (
      ModeInstall Mode = iota
      ModeReadFile
      ModeReadDir
  )

  type Invocation struct {
      Mode Mode
      Path string
  }

  func ParseArgs(args []string) (Invocation, error)
  ```

- [ ] **Step 1: Initialize the module and dependencies**

```bash
go mod init openmd
go get github.com/yuin/goldmark@latest
go get github.com/alecthomas/chroma/v2@latest
go get github.com/yuin/goldmark-highlighting/v2@latest
go get golang.org/x/sys@latest
```

- [ ] **Step 2: Write the failing test**

```go
package app

import "testing"

func TestParseArgs(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        want    Invocation
        wantErr bool
    }{
        {"no args", nil, Invocation{Mode: ModeInstall}, false},
        {"file", []string{`C:\docs\README.md`}, Invocation{Mode: ModeReadFile, Path: `C:\docs\README.md`}, false},
        {"dir", []string{"--dir", `C:\docs with space`}, Invocation{Mode: ModeReadDir, Path: `C:\docs with space`}, false},
        {"missing dir value", []string{"--dir"}, Invocation{}, true},
        {"unknown flag", []string{"--bogus"}, Invocation{}, true},
        {"two paths", []string{`C:\a.md`, `C:\b.md`}, Invocation{}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseArgs(tt.args)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Fatalf("got %#v, want %#v", got, tt.want)
            }
        })
    }
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/app -run TestParseArgs -v`  
Expected: FAIL because `ParseArgs` is not defined.

- [ ] **Step 4: Implement `ParseArgs` and a minimal `main.go`**

Rules:
- Empty args → `ModeInstall`.
- Exactly one non-flag argument → `ModeReadFile`.
- `--dir` plus exactly one value → `ModeReadDir`.
- Missing values, unknown flags, or more than one path → error.

`main.go` only needs to compile in this task; real dispatch is added in Task 9.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/app -run TestParseArgs -v`  
Expected: PASS.

- [ ] **Step 6: Checkpoint**

Run: `go test ./...`  
Expected: PASS for all packages created so far.

---

### Task 2: Markdown rendering, title extraction, and TOC

**Files:**
- Create: `internal/render/document.go`
- Create: `internal/render/toc.go`
- Test: `internal/render/document_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces:
  ```go
  package render

  type Heading struct {
      Level int
      Text  string
      ID    string
  }

  type Document struct {
      Path    string
      RelPath string
      ID      string
      Title   string
      HTML    template.HTML
      TOC     []Heading
  }

  type RenderOptions struct {
      IDPrefix  string
      KnownDocs map[string]string
  }

  func RenderMarkdown(source []byte, sourcePath string, opts RenderOptions) (*Document, error)
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestRenderMarkdownGFMAndTOC(t *testing.T) {
    src := []byte("# 标题\n\n## 用法\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n")
    doc, err := RenderMarkdown(src, `C:\docs\README.md`, RenderOptions{})
    if err != nil { t.Fatal(err) }
    if doc.Title != "标题" { t.Fatalf("title = %q", doc.Title) }
    if len(doc.TOC) != 2 { t.Fatalf("toc len = %d", len(doc.TOC)) }
    if doc.TOC[0].ID == "" || doc.TOC[1].ID == "" { t.Fatal("missing heading IDs") }
    html := string(doc.HTML)
    if !strings.Contains(html, "<table>") { t.Fatal("GFM table missing") }
    if !strings.Contains(html, `type="checkbox"`) { t.Fatal("task list missing") }
}

func TestRenderMarkdownRawHTMLIsOmitted(t *testing.T) {
    doc, err := RenderMarkdown([]byte("<script>alert(1)</script>\n\nhello"), "note.md", RenderOptions{})
    if err != nil { t.Fatal(err) }
    if strings.Contains(string(doc.HTML), "<script>") { t.Fatal("raw HTML was rendered") }
    if !strings.Contains(string(doc.HTML), "hello") { t.Fatal("normal text missing") }
}

func TestRenderMarkdownTitleFallbackAndHighlighting(t *testing.T) {
    doc, err := RenderMarkdown([]byte("```go\nfunc main() {}\n```\n"), `C:\docs\no-title.md`, RenderOptions{})
    if err != nil { t.Fatal(err) }
    if doc.Title != "no-title" { t.Fatalf("title = %q", doc.Title) }
    if !strings.Contains(string(doc.HTML), "chroma") { t.Fatal("Chroma classes missing") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/render -run 'TestRenderMarkdown' -v`  
Expected: FAIL because `RenderMarkdown` is not defined.

- [ ] **Step 3: Implement the Markdown pipeline**

- Configure goldmark with GFM, footnotes, auto heading IDs, and Chroma highlighting using CSS classes.
- Walk the AST after parsing to collect h1–h3 headings for the TOC and the first level-1 heading for the title.
- Apply `RenderOptions.IDPrefix` to every heading ID and its matching TOC link, so directory pages can embed multiple documents without duplicate IDs.
- Fall back to the filename without extension when no h1 exists.
- Render only the document body into `Document.HTML`; the full page wrapper is Task 4.
- Leave `KnownDocs` unused in this task; Task 3 wires URL rewriting.
- Keep raw HTML disabled.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/render -run 'TestRenderMarkdown' -v`  
Expected: PASS.

- [ ] **Step 5: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 3: Local URL rewriting for images and links

**Files:**
- Create: `internal/render/url.go`
- Test: `internal/render/url_test.go`
- Modify: `internal/render/document.go`

**Interfaces:**
- Consumes: `RenderMarkdown` from Task 2.
- Produces:
  ```go
  package render

  type URLKind int
  const (
      LinkURL URLKind = iota
      ImageURL
  )

  func RewriteURL(raw, baseDir string, kind URLKind, knownDocs map[string]string) (string, error)
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestRewriteLocalImagePaths(t *testing.T) {
    base := `C:\docs with space\notes`
    got, err := RewriteURL("images/a b.png", base, ImageURL, nil)
    if err != nil { t.Fatal(err) }
    if !strings.HasPrefix(got, "file:///C:/docs%20with%20space/notes/images/") {
        t.Fatalf("got %q", got)
    }
    if !strings.Contains(got, "a%20b.png") {
        t.Fatalf("space/unicode encoding missing: %q", got)
    }
}

func TestRewriteChinesePathAndUNC(t *testing.T) {
    got, err := RewriteURL("图片/图 1.png", `C:\文档`, ImageURL, nil)
    if err != nil { t.Fatal(err) }
    if !strings.Contains(got, "%E5%9B%BE%E7%89%87") || !strings.Contains(got, "%E5%9B%BE%201.png") {
        t.Fatalf("unicode encoding missing: %q", got)
    }
    unc, err := RewriteURL(`\\server\share\a.png`, `C:\docs`, ImageURL, nil)
    if err != nil { t.Fatal(err) }
    if unc != "file://server/share/a.png" { t.Fatalf("unc = %q", unc) }
}

func TestRewriteRemoteAnchorAndDataURLsAreUnchanged(t *testing.T) {
    for _, raw := range []string{"https://example.com/a.png", "#section", "data:image/png;base64,AAAA"} {
        got, err := RewriteURL(raw, `C:\docs`, ImageURL, nil)
        if err != nil { t.Fatal(err) }
        if got != raw { t.Fatalf("raw %q became %q", raw, got) }
    }
}

func TestRewriteDangerousSchemes(t *testing.T) {
    for _, raw := range []string{"javascript:alert(1)", "vbscript:msgbox(1)"} {
        got, err := RewriteURL(raw, `C:\docs`, LinkURL, nil)
        if err != nil { t.Fatal(err) }
        if got != "" { t.Fatalf("got %q, want empty", got) }
    }
}

func TestRewriteKnownDocumentLink(t *testing.T) {
    known := map[string]string{`c:\docs\other.md`: "doc-1"}
    got, err := RewriteURL("other.md#usage", `C:\docs`, LinkURL, known)
    if err != nil { t.Fatal(err) }
    if got != "#doc=doc-1&anchor=usage" { t.Fatalf("got %q", got) }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/render -run 'TestRewrite' -v`  
Expected: FAIL because `RewriteURL` is not defined.

- [ ] **Step 3: Implement `RewriteURL`**

- Leave `http`, `https`, `mailto`, `tel`, `file`, fragment-only, and local `data:image/*` values unchanged for images.
- Resolve local paths relative to `baseDir`; make them absolute; convert to `file:///C:/...` with `net/url` so spaces, Chinese, `#`, and `%` are encoded correctly.
- Handle UNC paths as `file://server/share/...`.
- For links, map a resolved path found in `knownDocs` to `#doc=<id>`; when a fragment exists, append `&anchor=<fragment>`.
- Return `""` for `javascript:` and `vbscript:`.
- Key `knownDocs` by lower-cased, cleaned absolute paths because Windows file paths are case-insensitive.

- [ ] **Step 4: Integrate rewriting into `RenderMarkdown`**

Walk `ast.Image` and `ast.Link` nodes and replace their destinations with `RewriteURL` results. If a destination becomes `""`, leave an empty destination so no dangerous handler is emitted.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/render -v`  
Expected: PASS, including all Task 2 tests.

- [ ] **Step 6: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 4: Embedded page template, CSS, JS, and single-file page

**Files:**
- Create: `internal/render/page.go`
- Create: `internal/render/assets.go`
- Create: `internal/render/web/page.html`
- Create: `internal/render/web/reader.css`
- Create: `internal/render/web/reader.js`
- Test: `internal/render/page_test.go`

**Interfaces:**
- Consumes: `RenderMarkdown` and `Document` from Tasks 2–3.
- Produces:
  ```go
  package render

  type Page struct {
      Mode      string // "file" or "dir"
      Title     string
      Active    int
      Documents []Document
      Warnings  []string
  }

  func BuildFilePage(filePath string) (*Page, error)
  func RenderPage(page *Page) ([]byte, error)
  func HighlightCSS() (light, dark string)
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestBuildFilePage(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "note.md")
    os.WriteFile(path, []byte("# Hello\n\nBody\n"), 0644)
    page, err := BuildFilePage(path)
    if err != nil { t.Fatal(err) }
    if page.Mode != "file" { t.Fatalf("mode = %q", page.Mode) }
    if page.Title != "Hello" { t.Fatalf("title = %q", page.Title) }
    if len(page.Documents) != 1 || page.Active != 0 { t.Fatalf("docs = %#v active = %d", page.Documents, page.Active) }
}

func TestRenderPageContainsEmbeddedUIAndCSP(t *testing.T) {
    page := &Page{Mode: "file", Title: "T", Documents: []Document{{ID: "doc-0", Title: "T", HTML: template.HTML("<p>hi</p>"), TOC: []Heading{{Level: 1, Text: "T", ID: "doc-0-t"}}}}}
    out, err := RenderPage(page)
    if err != nil { t.Fatal(err) }
    s := string(out)
    for _, want := range []string{"<!doctype html>", "Content-Security-Policy", "<style>", "<script>", `data-doc-id="doc-0"`, `class="toc-template"`} {
        if !strings.Contains(s, want) { t.Fatalf("missing %q", want) }
    }
    if strings.Contains(s, "<link rel=") || strings.Contains(s, "script src=") { t.Fatal("external asset referenced") }
}

func TestRenderPageEscapesTitle(t *testing.T) {
    page := &Page{Mode: "file", Title: `<script>alert(1)</script>`, Documents: []Document{{ID: "doc-0", Title: "x"}}}
    out, err := RenderPage(page)
    if err != nil { t.Fatal(err) }
    if strings.Contains(string(out), "<script>alert(1)</script>") { t.Fatal("title was not escaped") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/render -run 'TestBuildFilePage|TestRenderPage' -v`  
Expected: FAIL because the page API is not defined.

- [ ] **Step 3: Create the page template and assets**

Template requirements:
- Chinese `lang="zh-CN"`, UTF-8, viewport, CSP meta.
- Top bar with title and mobile sidebar toggle.
- A top-of-page warning banner when `Page.Warnings` is non-empty.
- Directory mode sidebar has two tabs: “文件列表” and “大纲”, defaulting to “文件列表”; single-file mode shows only “大纲”.
- Article content uses vendored `github-markdown-css`; custom CSS owns the page shell.
- Sidebar contains `#doc-list` (directory mode only) and `#toc`.
- Main content contains one `<section data-doc-id="...">` per document.
- Each section contains a `<template class="toc-template">` with its TOC links.
- No external stylesheet, script, font, or iframe.

CSS requirements:
- GitHub-like reading layout, max content width 820–920px, comfortable Chinese typography.
- Fixed sidebar around 260–300px; collapsible under a narrow breakpoint.
- Light/dark variables via `prefers-color-scheme`.
- Code, tables, blockquotes, task lists, images, and scrollbars styled.
- Include `HighlightCSS()` output for both light and dark Chroma styles.

JS requirements:
- Clone the active document’s TOC template into `#toc`.
- Directory mode: clicking a document button switches visible section, updates title, hash, and TOC.
- Parse `#doc=<id>` and `#doc=<id>&anchor=<anchor>`; switch document and scroll to the matching prefixed heading.
- Scrollspy highlights the active TOC entry.
- Mobile toggle shows/hides the sidebar.
- No external libraries.

- [ ] **Step 4: Implement `BuildFilePage`, `RenderPage`, and `HighlightCSS`**

- `BuildFilePage` reads the file, calls `RenderMarkdown` with `RenderOptions{IDPrefix: "doc-0"}`, sets `ID: "doc-0"`, and uses the document title for the page title.
- Prefix heading IDs with the document ID during rendering so multiple documents can share one DOM. For the single-file page this produces IDs like `doc-0-usage`.
- `RenderPage` executes the embedded template with embedded CSS/JS.
- `HighlightCSS` uses Chroma’s `github` and `github-dark` styles and returns CSS scoped by `prefers-color-scheme`.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/render -v`  
Expected: PASS.

- [ ] **Step 6: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 5: Directory scan and multi-document page

**Files:**
- Create: `internal/scan/scan.go`
- Test: `internal/scan/scan_test.go`
- Create: `internal/render/directory.go`
- Test: `internal/render/directory_test.go`
- Modify: `internal/render/page.go`

**Interfaces:**
- Consumes: `RenderMarkdown`, `Document`, `Page`, and `Template` from earlier tasks.
- Produces:
  ```go
  package scan

  type Limits struct {
      MaxDocs       int
      MaxTotalBytes int64
  }

  var DefaultLimits = Limits{MaxDocs: 500, MaxTotalBytes: 25 << 20}

  type Document struct {
      Path string
      Name string
      Size int64
  }

  type Result struct {
      Documents []Document
      Warnings  []string
      Truncated bool
  }

  var ErrNoMarkdown = errors.New("no markdown files")

  func Directory(dir string, limits Limits) (*Result, error)
  ```

  ```go
  package render

  func BuildDirPage(dirPath string) (*Page, error)
  ```

- [ ] **Step 1: Write the failing scan tests**

```go
func TestDirectoryScansCurrentLevelOnly(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "b.MD"), []byte("b"), 0644)
    os.WriteFile(filepath.Join(dir, "a.markdown"), []byte("a"), 0644)
    os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0644)
    os.Mkdir(filepath.Join(dir, "sub"), 0755)
    os.WriteFile(filepath.Join(dir, "sub", "nested.md"), []byte("n"), 0644)

    got, err := Directory(dir, DefaultLimits)
    if err != nil { t.Fatal(err) }
    if len(got.Documents) != 2 { t.Fatalf("docs = %#v", got.Documents) }
    if got.Documents[0].Name != "a.markdown" || got.Documents[1].Name != "b.MD" {
        t.Fatalf("order = %#v", got.Documents)
    }
}

func TestDirectoryNaturalSort(t *testing.T) {
    dir := t.TempDir()
    for _, name := range []string{"file10.md", "file2.md", "file1.md"} {
        os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644)
    }
    got, err := Directory(dir, DefaultLimits)
    if err != nil { t.Fatal(err) }
    names := []string{got.Documents[0].Name, got.Documents[1].Name, got.Documents[2].Name}
    want := []string{"file1.md", "file2.md", "file10.md"}
    for i := range want {
        if names[i] != want[i] { t.Fatalf("names = %#v", names) }
    }
}

func TestDirectoryLimits(t *testing.T) {
    dir := t.TempDir()
    for _, name := range []string{"a.md", "b.md", "c.md"} {
        os.WriteFile(filepath.Join(dir, name), []byte("12345"), 0644)
    }
    got, err := Directory(dir, Limits{MaxDocs: 2, MaxTotalBytes: 100})
    if err != nil { t.Fatal(err) }
    if len(got.Documents) != 2 || !got.Truncated { t.Fatalf("docs=%d truncated=%v", len(got.Documents), got.Truncated) }
}

func TestDirectoryTotalBytesLimit(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "a.md"), []byte("12345"), 0644)
    os.WriteFile(filepath.Join(dir, "b.md"), []byte("12345"), 0644)
    got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 5})
    if err != nil { t.Fatal(err) }
    if len(got.Documents) != 1 || !got.Truncated { t.Fatalf("docs=%d truncated=%v", len(got.Documents), got.Truncated) }
}

func TestDirectoryEmpty(t *testing.T) {
    _, err := Directory(t.TempDir(), DefaultLimits)
    if !errors.Is(err, ErrNoMarkdown) { t.Fatalf("err = %v", err) }
}
```

- [ ] **Step 2: Run the scan tests to verify they fail**

Run: `go test ./internal/scan -v`  
Expected: FAIL because `Directory` is not defined.

- [ ] **Step 3: Implement `scan.Directory`**

- Read only the current directory.
- Accept `.md` and `.markdown` case-insensitively.
- Ignore subdirectories, including directories whose names end in `.md`.
- Sort by filename with a natural comparison, so `file2.md` comes before `file10.md`.
- Enforce `MaxDocs` and `MaxTotalBytes`; set `Truncated` when a limit stops the scan.
- Return `ErrNoMarkdown` when no usable files remain.
- Add warnings for entries that cannot be statted and continue.

- [ ] **Step 4: Run the scan tests to verify they pass**

Run: `go test ./internal/scan -v`  
Expected: PASS.

- [ ] **Step 5: Write the failing directory-page tests**

```go
func TestBuildDirPageAndCrossDocumentLink(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n\n[B](b.md)\n"), 0644)
    os.WriteFile(filepath.Join(dir, "b.md"), []byte("# B\n\n## Usage\n"), 0644)

    page, err := BuildDirPage(dir)
    if err != nil { t.Fatal(err) }
    if page.Mode != "dir" || len(page.Documents) != 2 { t.Fatalf("page = %#v", page) }
    if !strings.Contains(string(page.Documents[0].HTML), "#doc=doc-1") {
        t.Fatalf("cross-document link not rewritten: %s", page.Documents[0].HTML)
    }
    if !strings.Contains(string(page.Documents[1].HTML), `id="doc-1-usage"`) {
        t.Fatalf("heading ID not prefixed: %s", page.Documents[1].HTML)
    }
}

func TestBuildDirPageKeepsGoingAfterUnreadableDocument(t *testing.T) {
    docs := []scan.Document{
        {Path: filepath.Join(t.TempDir(), "missing.md"), Name: "missing.md"},
        {Path: writeTempMD(t, "# OK"), Name: "ok.md"},
    }
    page, err := buildDirPage(docs)
    if err != nil { t.Fatal(err) }
    if len(page.Documents) != 2 { t.Fatalf("docs = %d", len(page.Documents)) }
    if !strings.Contains(string(page.Documents[0].HTML), "无法读取") { t.Fatal("missing error placeholder") }
}

func TestRenderPageShowsScanWarnings(t *testing.T) {
    page := &Page{Mode: "dir", Title: "docs", Warnings: []string{"只加载了前 500 篇"}, Documents: []Document{{ID: "doc-0", Title: "A"}}}
    out, err := RenderPage(page)
    if err != nil { t.Fatal(err) }
    if !strings.Contains(string(out), "只加载了前 500 篇") { t.Fatal("warning not shown") }
}
```

- [ ] **Step 6: Run the directory-page tests to verify they fail**

Run: `go test ./internal/render -run 'TestBuildDirPage' -v`  
Expected: FAIL because `BuildDirPage` is not defined.

- [ ] **Step 7: Implement `BuildDirPage` and `buildDirPage`**

- `BuildDirPage` scans with `scan.DefaultLimits`, builds a map of lower-cased absolute paths to `doc-N` IDs, and renders each document with `RenderOptions{IDPrefix: "doc-N", KnownDocs: knownDocs}`.
- `buildDirPage(docs []scan.Document)` is the injectable inner function for testing unreadable files.
- Set each document’s `RelPath` to its current-level filename for the sidebar list.
- A read failure produces a placeholder document titled with the filename and body text containing `无法读取`.
- The page title is the directory name.
- Warnings from the scan are stored on `Page.Warnings`; when the scan is truncated, add a warning such as `目录内容过多，只加载了前 N 篇`.

- [ ] **Step 8: Run the tests to verify they pass**

Run: `go test ./internal/scan ./internal/render -v`  
Expected: PASS.

- [ ] **Step 9: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 6: Temporary run directory manager

**Files:**
- Create: `internal/tempdir/manager.go`
- Test: `internal/tempdir/manager_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  package tempdir

  type Manager struct {
      BaseDir string
      Now     func() time.Time
      MaxAge  time.Duration
      MaxRuns int
  }

  func New(baseDir string) *Manager
  func (m *Manager) RunDir() (string, error)
  func (m *Manager) Cleanup() error
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestRunDirCreatesUniqueDirectories(t *testing.T) {
    m := New(t.TempDir())
    a, err := m.RunDir(); if err != nil { t.Fatal(err) }
    b, err := m.RunDir(); if err != nil { t.Fatal(err) }
    if a == b { t.Fatal("run dirs are identical") }
    for _, p := range []string{a, b} {
        if filepath.Dir(p) != m.BaseDir { t.Fatalf("%q not under %q", p, m.BaseDir) }
        if !strings.HasPrefix(filepath.Base(p), "run-") { t.Fatalf("bad name %q", p) }
    }
}

func TestCleanupRemovesOldAndKeepsNew(t *testing.T) {
    base := t.TempDir()
    now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
    m := New(base)
    m.Now = func() time.Time { return now }
    m.MaxAge = 7 * 24 * time.Hour
    old := filepath.Join(base, "run-old"); os.Mkdir(old, 0755)
    fresh := filepath.Join(base, "run-fresh"); os.Mkdir(fresh, 0755)
    os.Chtimes(old, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour))
    os.Chtimes(fresh, now, now)
    if err := m.Cleanup(); err != nil { t.Fatal(err) }
    if _, err := os.Stat(old); !os.IsNotExist(err) { t.Fatal("old run dir not removed") }
    if _, err := os.Stat(fresh); err != nil { t.Fatal("fresh run dir removed") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/tempdir -v`  
Expected: FAIL because `Manager` is not defined.

- [ ] **Step 3: Implement `Manager`**

- `New` sets `MaxAge = 7 * 24 * time.Hour`, `MaxRuns = 50`, and `Now = time.Now`.
- `RunDir` creates `BaseDir` if needed, then `os.MkdirTemp(BaseDir, "run-"+timestamp+"-*")`.
- `Cleanup` considers only entries whose names start with `run-` and that are directories.
- Remove entries older than `MaxAge`; then keep only the newest `MaxRuns` by mod time.
- Ignore individual removal failures; return only a base-directory read/creation error.
- Add a test for the 50-directory cap using 55 directories with distinct mod times.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/tempdir -v`  
Expected: PASS.

- [ ] **Step 5: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 7: Default-browser opener

**Files:**
- Create: `internal/browser/open.go`
- Test: `internal/browser/open_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  package browser

  type Opener interface {
      Open(path string) error
  }

  type CommandRunner func(name string, args ...string) error

  type WindowsOpener struct {
      Run CommandRunner
  }

  func NewWindowsOpener() *WindowsOpener
  ```

- [ ] **Step 1: Write the failing tests**

```go
func TestWindowsOpenerUsesRundll32(t *testing.T) {
    var calls [][]string
    o := &WindowsOpener{Run: func(name string, args ...string) error {
        calls = append(calls, append([]string{name}, args...))
        return nil
    }}
    if err := o.Open(`C:\temp\openmd\index.html`); err != nil { t.Fatal(err) }
    want := []string{"rundll32.exe", "url.dll,FileProtocolHandler", `C:\temp\openmd\index.html`}
    if !reflect.DeepEqual(calls[0], want) { t.Fatalf("calls = %#v", calls) }
}

func TestWindowsOpenerFallsBackToExplorer(t *testing.T) {
    calls := 0
    o := &WindowsOpener{Run: func(name string, args ...string) error {
        calls++
        if calls == 1 { return errors.New("rundll32 failed") }
        return nil
    }}
    if err := o.Open("x.html"); err != nil { t.Fatal(err) }
    if calls != 2 { t.Fatalf("calls = %d", calls) }
}

func TestWindowsOpenerReturnsErrorWhenBothFail(t *testing.T) {
    o := &WindowsOpener{Run: func(string, ...string) error { return errors.New("no browser") }}
    if err := o.Open("x.html"); err == nil { t.Fatal("want error") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/browser -v`  
Expected: FAIL because `WindowsOpener` is not defined.

- [ ] **Step 3: Implement `WindowsOpener`**

- Default `Run` starts `exec.Command(name, args...).Start()` and releases the process without waiting.
- Try `rundll32.exe url.dll,FileProtocolHandler <path>`.
- On failure, try `explorer.exe <path>`.
- Return the last error if both fail.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/browser -v`  
Expected: PASS.

- [ ] **Step 5: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 8: Registry association, self-install, and uninstall

**Files:**
- Create: `internal/install/assoc.go`
- Create: `internal/install/registry_windows.go`
- Create: `internal/install/installer.go`
- Test: `internal/install/installer_test.go`
- Test: `internal/install/registry_windows_test.go`

**Interfaces:**
- Consumes: `HKCU` registry access and Windows shell APIs.
- Produces:
  ```go
  package install

  type Associator interface {
      Register(exePath string) error
      Unregister() error
  }

  type UninstallResult struct {
      NeedsHelper bool
  }

  type Installer struct {
      InstallDir string
      Assoc      Associator
      Notify     func()
      Spawn      func(exe, installDir string) error
  }

  func DefaultInstallDir() (string, error)
  func (i *Installer) InstalledExe() string
  func (i *Installer) Install(sourceExe string) error
  func (i *Installer) Uninstall(sourceExe string) (UninstallResult, error)
  func (i *Installer) SpawnUninstallHelper(sourceExe string) error
  ```
  Also:
  ```go
  type RegistryAssociator struct {
      ClassesRoot string // default: Software\Classes; injectable for tests
  }
  func NewRegistryAssociator() *RegistryAssociator
  ```

- [ ] **Step 1: Write the failing installer tests**

```go
func TestInstallCopiesExeAndRegisters(t *testing.T) {
    src := filepath.Join(t.TempDir(), "OpenMD.exe")
    os.WriteFile(src, []byte("binary"), 0644)
    installDir := filepath.Join(t.TempDir(), "OpenMD")
    fake := &fakeAssoc{}
    notified := 0
    in := &Installer{InstallDir: installDir, Assoc: fake, Notify: func() { notified++ }}
    if err := in.Install(src); err != nil { t.Fatal(err) }
    if _, err := os.Stat(in.InstalledExe()); err != nil { t.Fatal(err) }
    if fake.registeredPath != in.InstalledExe() { t.Fatalf("registered %q", fake.registeredPath) }
    if notified != 1 { t.Fatalf("notify = %d", notified) }
}

func TestInstallFromInstalledPathDoesNotOverwriteItself(t *testing.T) {
    installDir := filepath.Join(t.TempDir(), "OpenMD")
    in := &Installer{InstallDir: installDir, Assoc: &fakeAssoc{}, Notify: func() {}}
    os.MkdirAll(installDir, 0755)
    os.WriteFile(in.InstalledExe(), []byte("keep"), 0644)
    if err := in.Install(in.InstalledExe()); err != nil { t.Fatal(err) }
    data, _ := os.ReadFile(in.InstalledExe())
    if string(data) != "keep" { t.Fatalf("installed exe was overwritten: %q", data) }
}

func TestInstallAssociationFailureRollsBackFreshInstall(t *testing.T) {
    src := filepath.Join(t.TempDir(), "OpenMD.exe")
    os.WriteFile(src, []byte("binary"), 0644)
    installDir := filepath.Join(t.TempDir(), "OpenMD")
    in := &Installer{
        InstallDir: installDir,
        Assoc: &fakeAssoc{err: errors.New("access denied")},
        Notify: func() {},
    }
    if err := in.Install(src); err == nil { t.Fatal("want error") }
    if _, err := os.Stat(in.InstalledExe()); !os.IsNotExist(err) {
        t.Fatalf("failed install left an exe behind: %v", err)
    }
}

func TestSpawnUninstallHelperUsesInjectedSpawn(t *testing.T) {
    var gotExe, gotDir string
    in := &Installer{
        InstallDir: filepath.Join(t.TempDir(), "OpenMD"),
        Assoc: &fakeAssoc{},
        Notify: func() {},
        Spawn: func(exe, installDir string) error {
            gotExe, gotDir = exe, installDir
            return nil
        },
    }
    if err := in.SpawnUninstallHelper(`C:\Local\OpenMD.exe`); err != nil { t.Fatal(err) }
    if gotExe != `C:\Local\OpenMD.exe` || gotDir != in.InstallDir {
        t.Fatalf("spawn args = %q %q", gotExe, gotDir)
    }
}

func TestUninstallFromOtherLocationRemovesInstallDir(t *testing.T) {
    installDir := t.TempDir()
    in := &Installer{InstallDir: installDir, Assoc: &fakeAssoc{}, Notify: func() {}}
    os.WriteFile(in.InstalledExe(), []byte("x"), 0644)
    res, err := in.Uninstall(filepath.Join(t.TempDir(), "OpenMD.exe"))
    if err != nil { t.Fatal(err) }
    if res.NeedsHelper { t.Fatal("unexpected helper") }
    if _, err := os.Stat(installDir); !os.IsNotExist(err) { t.Fatal("install dir still exists") }
}

func TestUninstallFromInstalledPathNeedsHelper(t *testing.T) {
    installDir := t.TempDir()
    in := &Installer{InstallDir: installDir, Assoc: &fakeAssoc{}, Notify: func() {}}
    os.WriteFile(in.InstalledExe(), []byte("x"), 0644)
    res, err := in.Uninstall(in.InstalledExe())
    if err != nil { t.Fatal(err) }
    if !res.NeedsHelper { t.Fatal("want helper") }
    if _, err := os.Stat(in.InstalledExe()); err != nil { t.Fatal("installed exe removed too early") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/install -v`  
Expected: FAIL because the installer API is not defined.

- [ ] **Step 3: Implement the installer**

- `DefaultInstallDir` uses `%LOCALAPPDATA%` (fallback `os.UserCacheDir`) plus `Programs\OpenMD`.
- `Install` creates the install directory, copies the source exe to a temporary file beside the target, then renames it to `OpenMD.exe`.
- If the source path is already the installed exe, skip copying and only re-register.
- Copy failure or association failure returns an error; do not report success.
- On a fresh install, if association fails after copying, delete the copied exe so no half-install remains. On an update, keep a backup of the previous exe, restore it if association fails, and report the error.
- `Uninstall` calls `Assoc.Unregister` and removes `installDir\openmd-temp`.
- If the source is the installed exe, return `NeedsHelper=true` and leave the install directory for the helper.
- Otherwise remove the whole install directory.
- `SpawnUninstallHelper` calls the injected `Spawn(sourceExe, installDir)`. If `Spawn` is nil, use the real helper spawner, which copies the current exe to `%TEMP%` with a random name and starts it with `--uninstall-helper <installDir>`.
- The helper waits for the parent process to exit, deletes the install directory, then schedules deletion of its own temporary copy.
- The helper behavior itself is manual acceptance; unit-test the injected `Spawn` call path.

- [ ] **Step 4: Write the failing registry-preservation test**

Use a unique root under `HKCU\Software\OpenMDTest-<timestamp>` and clean it up with `defer`.

```go
func TestRegistryAssociatorPreservesOtherAssociations(t *testing.T) {
    root := fmt.Sprintf(`Software\OpenMDTest-%d`, time.Now().UnixNano())
    defer registry.CURRENT_USER.DeleteKey(root)
    k, _, err := registry.CreateKey(registry.CURRENT_USER, root+`\.md`, registry.ALL_ACCESS)
    if err != nil { t.Fatal(err) }
    k.SetStringValue("", "OtherApp.md")
    ow, _, err := registry.CreateKey(k, "OpenWithProgids", registry.ALL_ACCESS)
    if err != nil { t.Fatal(err) }
    ow.SetStringValue("Other.Prog", "")
    k.Close(); ow.Close()

    a := &RegistryAssociator{ClassesRoot: root}
    if err := a.Register(`C:\Tools\OpenMD.exe`); err != nil { t.Fatal(err) }

    // The Open With, file menu, and folder background entries must exist after install.
    for _, key := range []string{
        root + `\OpenMD.Markdown\shell\open\command`,
        root + `\.md\OpenWithProgids`,
        root + `\SystemFileAssociations\.md\shell\OpenMD\command`,
        root + `\Directory\shell\OpenMD\command`,
        root + `\Directory\Background\shell\OpenMD\command`,
    } {
        k, _, err := registry.OpenKey(registry.CURRENT_USER, key, registry.READ)
        if err != nil { t.Fatalf("missing key %q: %v", key, err) }
        k.Close()
    }

    if err := a.Unregister(); err != nil { t.Fatal(err) }

    k, _, err = registry.OpenKey(registry.CURRENT_USER, root+`\.md`, registry.READ)
    if err != nil { t.Fatal(err) }
    defer k.Close()
    v, _, err := k.GetStringValue("")
    if err != nil || v != "OtherApp.md" { t.Fatalf("default association changed: %q %v", v, err) }
    ow, _, err = registry.OpenKey(k, "OpenWithProgids", registry.READ)
    if err != nil { t.Fatal(err) }
    defer ow.Close()
    if _, _, err := ow.GetStringValue("Other.Prog"); err != nil { t.Fatal("other app's OpenWithProgids value removed") }
    if _, _, err := ow.GetStringValue("OpenMD.Markdown"); err == nil { t.Fatal("OpenMD value still present") }
}
```

- [ ] **Step 5: Run the registry test to verify it fails**

Run: `go test ./internal/install -run TestRegistryAssociatorPreservesOtherAssociations -v`  
Expected: FAIL because `RegistryAssociator` is not defined.

- [ ] **Step 6: Implement the real registry associator**

Register exactly the keys from the design:

- `OpenMD.Markdown` ProgID with `DefaultIcon` and `shell\open\command`.
- `Applications\OpenMD.exe` with `FriendlyAppName`, `shell\open\command`, and `SupportedTypes`.
- `.md` and `.markdown` `OpenWithProgids\OpenMD.Markdown`.
- `SystemFileAssociations\.md\shell\OpenMD` and `.markdown` equivalent.
- `Directory\shell\OpenMD` and `Directory\Background\shell\OpenMD`.
- Do not write the default value of `.md` or `.markdown`.

Unregister removes only those keys/values:
- Delete `OpenWithProgids\OpenMD.Markdown` values without deleting the parent key.
- Delete the OpenMD ProgID, Applications key, file-menu keys, and directory-menu keys.
- Preserve all unrelated values and keys.

- [ ] **Step 7: Implement `SHChangeNotify` notification**

Call it after successful register/unregister. Keep it behind `Notify` in `Installer` so tests can observe it without touching the real shell. If `Notify` is nil, use the real `SHChangeNotify` implementation.

- [ ] **Step 8: Run the installer and registry tests**

Run: `go test ./internal/install -v`  
Expected: PASS.

- [ ] **Step 9: Checkpoint**

Run: `go test ./...`  
Expected: PASS.

---

### Task 9: Installer dialog, app wiring, and main entry point

**Files:**
- Create: `internal/ui/dialog_windows.go`
- Create: `internal/app/app.go`
- Create: `internal/app/run.go`
- Test: `internal/app/run_test.go`
- Modify: `main.go`

**Interfaces:**
- Consumes: all earlier packages.
- Produces:
  ```go
  package ui

  type Choice int
  const (
      ChoiceCancel Choice = iota
      ChoiceInstall
      ChoiceUninstall
  )

  type Dialog interface {
      Installer(installedPath string, installed bool) Choice
      Info(message string)
      Error(message string)
  }

  type WindowsDialog struct{}
  ```

  ```go
  package app

  type Installer interface {
      Install(sourceExe string) error
      Uninstall(sourceExe string) (install.UninstallResult, error)
      SpawnUninstallHelper(sourceExe string) error
      InstalledExe() string
  }

  type App struct {
      Dialog     ui.Dialog
      Installer  Installer
      Temp       *tempdir.Manager
      Opener     browser.Opener
      Executable func() (string, error)
  }

  func (a *App) Run(inv Invocation) error
  func Main()
  ```

- [ ] **Step 1: Write the failing app tests**

```go
func TestRunReadFileRendersTempHTMLAndOpensBrowser(t *testing.T) {
    md := filepath.Join(t.TempDir(), "note.md")
    os.WriteFile(md, []byte("# Hello\n\n![x](pic.png)\n"), 0644)
    opener := &fakeOpener{}
    a := &App{
        Dialog: &fakeDialog{},
        Installer: &fakeInstaller{},
        Temp: tempdir.New(t.TempDir()),
        Opener: opener,
        Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil },
    }
    if err := a.Run(Invocation{Mode: ModeReadFile, Path: md}); err != nil { t.Fatal(err) }
    if opener.path == "" { t.Fatal("browser not opened") }
    if filepath.Base(opener.path) != "index.html" { t.Fatalf("opened %q", opener.path) }
    data, err := os.ReadFile(opener.path)
    if err != nil { t.Fatal(err) }
    if !strings.Contains(string(data), "Hello") { t.Fatal("rendered content missing") }
    entries, err := os.ReadDir(filepath.Dir(md))
    if err != nil { t.Fatal(err) }
    if len(entries) != 1 || entries[0].Name() != "note.md" {
        t.Fatalf("markdown directory was modified: %#v", entries)
    }
}

func TestRunInstallChoiceCallsInstaller(t *testing.T) {
    fi := &fakeInstaller{installedExe: `C:\Local\OpenMD.exe`}
    d := &fakeDialog{choice: ui.ChoiceInstall}
    a := &App{
        Dialog: d,
        Installer: fi,
        Temp: tempdir.New(t.TempDir()),
        Opener: &fakeOpener{},
        Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil },
    }
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if !fi.installCalled { t.Fatal("install not called") }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/app -run TestRun -v`  
Expected: FAIL because `App` is not defined.

- [ ] **Step 3: Implement the Windows dialog**

- Use `MessageBoxW` with `MB_YESNOCANCEL | MB_ICONQUESTION`.
- Map Yes → `ChoiceInstall`, No → `ChoiceUninstall`, Cancel/close → `ChoiceCancel`.
- Text must explain the mapping explicitly.
- Show success and error messages with `MB_ICONINFORMATION` / `MB_ICONERROR`.
- No console output.

- [ ] **Step 4: Implement `App.Run`**

- `ModeInstall`:
  - Resolve the current executable with `Executable`.
  - Determine installed state from `Installer.InstalledExe` existence.
  - Call `Dialog.Installer`.
  - Install → `Installer.Install(currentExe)`, then `Dialog.Info`.
  - Uninstall → `Installer.Uninstall(currentExe)`; if `NeedsHelper`, call `SpawnUninstallHelper`; then `Dialog.Info`. If cleanup fails or a residual path remains, show that path in the message.
- `ModeReadFile`:
  - `BuildFilePage(path)` → `RenderPage`.
- `ModeReadDir`:
  - `BuildDirPage(path)` → `RenderPage`.
- For reading modes:
  - Call `Temp.Cleanup()`.
  - `Temp.RunDir()`, write `index.html` atomically (write `.tmp`, rename).
  - `Opener.Open(htmlPath)`.
  - Return errors; `Main` renders them through `Dialog.Error`.
- If `Temp.RunDir()` fails because the install directory is not writable, create a second manager rooted at `%TEMP%\OpenMD` and retry. Do not use the Markdown directory.

- [ ] **Step 5: Implement `Main` and `main.go`**

- Build real dependencies: `ui.WindowsDialog`, `install.NewRegistryAssociator`, `install.DefaultInstallDir`, `tempdir.New(installDir+`\openmd-temp`)`, `browser.NewWindowsOpener`.
- Parse args; on error show a dialog and exit.
- Dispatch through `App.Run`; on error show a dialog.
- Support the internal helper mode from Task 8.
- `main.go` calls `app.Main()`.
- Keep `main.go` minimal so the GUI-subsystem build has no console window.

- [ ] **Step 6: Run the app tests**

Run: `go test ./internal/app -v`  
Expected: PASS.

- [ ] **Step 7: Build the GUI executable**

Run:
```bash
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```
Expected: `OpenMD.exe` is produced; launching it with no args opens the installer dialog.

- [ ] **Step 8: Checkpoint**

Run: `go test ./...` and `go vet ./...`  
Expected: both PASS.

---

### Task 10: Documentation, verification, and manual acceptance

**Files:**
- Create: `README.md`
- No code changes unless verification finds a defect.

**Interfaces:**
- Consumes: the built `OpenMD.exe`.
- Produces: user-facing usage notes and a manual acceptance record.

- [ ] **Step 1: Write `README.md`**

Include:
- What OpenMD is and what it is not (read-only, no server, no editing).
- Install: double-click the exe, choose install; original download can be deleted.
- Usage: right-click md, right-click folder/folder background.
- Uninstall: double-click an OpenMD.exe and choose uninstall.
- Temp location and cleanup.
- Chrome/Edge recommendation and Firefox image limitation.
- Build command and test command.

- [ ] **Step 2: Run the full automated verification**

Run:
```bash
go test ./... -count=1
go vet ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```
Expected: all commands succeed.

- [ ] **Step 3: Manually verify on Windows**

Use the checklist from the design doc. At minimum:

1. Place `OpenMD.exe` in `Downloads`, double-click, install.
2. Delete the `Downloads` copy; confirm right-click still works against the installed copy.
3. Right-click an `.md` file → “用 OpenMD 阅读”; confirm browser opens and content renders.
4. Open a document with a relative image in another directory; confirm the image loads in Chrome/Edge.
5. Test a path containing spaces and Chinese characters.
6. Right-click a folder and the folder background; confirm both open the directory page.
7. Switch documents, verify TOC updates and cross-document links switch sections.
8. Confirm duplicate headings in two documents jump to the correct document.
9. Confirm temp HTML exists only under the install directory.
10. Confirm `.md` default association is unchanged.
11. Uninstall; confirm menus are gone and the install directory is removed.
12. Try an empty directory and a missing file; confirm clear error dialogs.

- [ ] **Step 4: Report results**

Record which manual checks passed and any known limitation. If a check fails, fix the defect and rerun the relevant automated tests plus the failed manual check.

- [ ] **Step 5: Final checkpoint**

Run: `go test ./... -count=1`  
Expected: PASS.

Do not commit or push unless the user explicitly asks.
