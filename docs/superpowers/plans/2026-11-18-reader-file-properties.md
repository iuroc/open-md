# Reader File Properties Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the current document's absolute path and directory to the existing information panel, plus a copy-path action and a circle-plus-i information icon.

**Architecture:** Go already carries each document's source path in `Document.Path`; extend it with `Dir`, normalize single-file paths to absolute, expose both on each document section as data attributes, and let the existing panel/JS keep the visible path in sync when documents switch. The panel stays a single popover; no new entry point.

**Tech Stack:** Go 1.26, `html/template`, vanilla JavaScript/CSS, goldmark.

**Spec:** `docs/superpowers/specs/2026-11-18-reader-file-properties-design.md`

## Global Constraints

- Windows paths; display absolute paths via `filepath.Abs` + `filepath.Clean`.
- Do not resolve symlinks; the displayed path is the absolute path OpenMD opened.
- Keep one panel and one topbar entry; no second properties button.
- Panel title and accessible label become `信息与设置`.
- Preserve existing themes, author, version, build time, and release link behavior.
- No new Go or JavaScript dependencies.

## Review Focus

- Long, spaced, or non-ASCII paths must wrap inside the 340px panel and remain selectable.
- Switching documents must update both the path and directory immediately, and clear any stale copy status.
- A relative single-file argument must still display an absolute path.
- Clipboard access may be blocked on `file://`; the fallback must work or show a clear failure.
- Linked documents and error placeholder documents must still show their own path and directory.

---

### Task 1: Document path data and section attributes

**Files:**
- Modify: `internal/render/document.go:26-34`
- Modify: `internal/render/links.go:35-72`
- Modify: `internal/render/page.go:44-64,100-125`
- Modify: `internal/render/web/page.html:138`
- Test: `internal/render/page_test.go`
- Test: `internal/render/links_test.go`

**Interfaces:**
- Consumes: existing `Document.Path`, `sourceDoc.Path`, `Page.Active`.
- Produces:
  - `Document.Dir string`
  - `makeAbsolute(path string) string`
  - `pageData.ActivePath string`, `pageData.ActiveDir string`
  - each `.doc-section` gets `data-path` and `data-dir`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/render/page_test.go`:

```go
func TestBuildFilePageNormalizesPathAndSetsDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("# Title\n"), 0644); err != nil {
		t.Fatal(err)
	}
	page, err := BuildFilePage(path)
	if err != nil {
		t.Fatal(err)
	}
	if page.Documents[0].Path != path {
		t.Fatalf("path = %q, want %q", page.Documents[0].Path, path)
	}
	if page.Documents[0].Dir != dir {
		t.Fatalf("dir = %q, want %q", page.Documents[0].Dir, dir)
	}
}

func TestMakeAbsolute(t *testing.T) {
	got := makeAbsolute("note.md")
	if !filepath.IsAbs(got) {
		t.Fatalf("makeAbsolute returned relative path %q", got)
	}
}

func TestRenderPageEmitsDocumentPathAttributes(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "A",
		Documents: []Document{{
			ID:    "doc-0",
			Title: "A",
			Path:  `C:\文档\a b.md`,
			Dir:   `C:\文档`,
		}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	for _, want := range []string{`data-path="C:\文档\a b.md"`, `data-dir="C:\文档"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in %s", want, html)
		}
	}
}
```

Add to `internal/render/links_test.go`:

```go
func TestBuildFilePageLinkedDocumentHasDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("[Other](other.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if page.Documents[1].Dir != dir {
		t.Fatalf("linked dir = %q, want %q", page.Documents[1].Dir, dir)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/render -run 'TestBuildFilePageNormalizesPathAndSetsDir|TestMakeAbsolute|TestRenderPageEmitsDocumentPathAttributes|TestBuildFilePageLinkedDocumentHasDir'`

Expected: FAIL because `Dir`, `makeAbsolute`, and the data attributes do not exist.

- [ ] **Step 3: Implement the data changes**

In `document.go`, add `Dir string` to `Document`.

In `makeAbsolute` (new helper in `page.go`):

```go
func makeAbsolute(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}
```

In `BuildFilePage`, normalize the input before stat:

```go
filePath = makeAbsolute(filePath)
```

In `buildPageFromSources` (`links.go`), after `doc.RelPath = sourceDoc.RelPath`, set:

```go
doc.Dir = filepath.Dir(sourceDoc.Path)
```

In `errorDocumentForSource`, include:

```go
Dir: filepath.Dir(doc.Path),
```

In `pageData`, add `ActivePath string` and `ActiveDir string`; in `RenderPage`, fill them from the active document (empty when no document exists).

In `page.html`, add attributes to the section tag:

```html
data-path="{{$doc.Path}}" data-dir="{{$doc.Dir}}"
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/render -run 'TestBuildFilePageNormalizesPathAndSetsDir|TestMakeAbsolute|TestRenderPageEmitsDocumentPathAttributes|TestBuildFilePageLinkedDocumentHasDir'`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/document.go internal/render/links.go internal/render/page.go internal/render/web/page.html internal/render/page_test.go internal/render/links_test.go
git commit -m "feat: expose document path data to the reader"
```

### Task 2: Panel markup, information icon, and styles

**Files:**
- Modify: `internal/render/web/page.html:86-125`
- Modify: `internal/render/web/reader.css:418-593`
- Test: `internal/render/page_test.go`

**Interfaces:**
- Consumes: `pageData.ActivePath`, `pageData.ActiveDir`, section `data-path` / `data-dir` from Task 1.
- Produces: DOM IDs `file-path`, `file-dir`, `copy-path`, `copy-status`; panel label/title `信息与设置`; info icon.

- [ ] **Step 1: Write the failing test**

Add to `internal/render/page_test.go`:

```go
func TestRenderPageShowsFilePropertiesPanel(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "A",
		Documents: []Document{{
			ID:    "doc-0",
			Title: "A",
			Path:  `C:\docs\a.md`,
			Dir:   `C:\docs`,
		}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	for _, want := range []string{
		`id="file-path"`,
		`id="file-dir"`,
		`id="copy-path"`,
		`id="copy-status"`,
		`信息与设置`,
		`aria-label="信息与设置"`,
		`C:\docs\a.md`,
		`C:\docs`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in panel: %s", want, html)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render -run TestRenderPageShowsFilePropertiesPanel`

Expected: FAIL because the panel markup does not exist.

- [ ] **Step 3: Implement the panel and icon**

Change the settings button label/title/panel title to `信息与设置`. Replace the gear SVG with:

```html
<circle cx="10" cy="10" r="7"></circle>
<path d="M10 9v5"></path>
<circle cx="10" cy="6.3" r="0.9" fill="currentColor" stroke="none"></circle>
```

Insert a file group before the color group:

```html
<div class="settings-group settings-file">
  <div class="settings-label">文件</div>
  <dl class="about-meta file-meta">
    <div><dt>当前文档</dt><dd id="file-path" class="file-path">{{.ActivePath}}</dd></div>
    <div><dt>所在目录</dt><dd id="file-dir" class="file-path">{{.ActiveDir}}</dd></div>
  </dl>
  <button type="button" id="copy-path" class="copy-path" aria-label="复制当前文档路径">复制路径</button>
  <span id="copy-status" class="copy-status" role="status" aria-live="polite"></span>
</div>
```

Add CSS:

```css
.file-path {
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-all;
  user-select: text;
  font-family: var(--font-mono);
  font-size: 11px;
}

.copy-path {
  margin-top: 8px;
  padding: 5px 9px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: transparent;
  color: var(--fg-muted);
  cursor: pointer;
  font: inherit;
  font-size: 12px;
}

.copy-path:hover {
  background: var(--bg-hover);
  color: var(--fg);
}

.copy-status {
  margin-left: 8px;
  color: var(--fg-muted);
  font-size: 12px;
}
```

Include `#copy-path` in the existing `:focus-visible` rule.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render -run TestRenderPageShowsFilePropertiesPanel`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/render/web/page.html internal/render/web/reader.css internal/render/page_test.go
git commit -m "feat: add file properties to the information panel"
```

### Task 3: Keep the panel path in sync and copy it

**Files:**
- Modify: `internal/render/web/reader.js:14-21,215-258,335-365`

**Interfaces:**
- Consumes: `#file-path`, `#file-dir`, `#copy-path`, `#copy-status`; section `data-path` / `data-dir`.
- Produces: `showDoc` updates path/dir; copy button writes the current path.

- [ ] **Step 1: Make a two-document fixture and verify the current behavior is wrong**

Create a temporary directory with `a.md` and `b.md`, open it with OpenMD, and append a probe script that clicks `[data-doc-target="doc-1"]` and reads `#file-path`. Before the JS change, `#file-path` should still show the first document. Keep the probe command and expected failure in the task notes.

- [ ] **Step 2: Implement the JS**

Add element references:

```js
var filePathEl = document.getElementById("file-path");
var fileDirEl = document.getElementById("file-dir");
var copyPathButton = document.getElementById("copy-path");
var copyStatusEl = document.getElementById("copy-status");
var currentPath = "";
```

In `showDoc`, after updating `document.title`:

```js
currentPath = section.getAttribute("data-path") || "";
if (filePathEl) filePathEl.textContent = currentPath;
if (fileDirEl) fileDirEl.textContent = section.getAttribute("data-dir") || "";
setCopyStatus("");
```

Implement:

```js
function setCopyStatus(text) {
  if (copyStatusEl) copyStatusEl.textContent = text;
}

function fallbackCopy(text) {
  var area = document.createElement("textarea");
  area.value = text;
  area.setAttribute("readonly", "");
  area.style.position = "fixed";
  area.style.opacity = "0";
  document.body.appendChild(area);
  area.select();
  var ok = false;
  try { ok = document.execCommand("copy"); } catch (err) { ok = false; }
  document.body.removeChild(area);
  return ok;
}

function copyCurrentPath() {
  if (!currentPath) {
    setCopyStatus("没有可复制的路径");
    return;
  }
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(currentPath).then(
      function () { setCopyStatus("已复制"); },
      function () { setCopyStatus(fallbackCopy(currentPath) ? "已复制" : "复制失败"); }
    );
    return;
  }
  setCopyStatus(fallbackCopy(currentPath) ? "已复制" : "复制失败");
}
```

Attach:

```js
if (copyPathButton) {
  copyPathButton.addEventListener("click", copyCurrentPath);
}
```

- [ ] **Step 3: Run the probe and verify the path updates**

Click the second document in the fixture. Expected: `#file-path` shows the second document's absolute path, `#file-dir` shows its parent directory, and `#copy-status` is empty.

- [ ] **Step 4: Verify copy behavior**

In Edge/Chrome, click `复制路径`, paste into a text field, and confirm the full path. If clipboard access is blocked, confirm the fallback still copies or shows `复制失败`.

- [ ] **Step 5: Commit**

```bash
git add internal/render/web/reader.js
git commit -m "feat: sync and copy the active document path"
```

### Task 4: Documentation and full verification

**Files:**
- Modify: `README.md`
- Verify: `go test ./...`, `go vet ./...`, `build.bat`

- [ ] **Step 1: Update README**

Add a feature bullet:

```markdown
- 顶栏信息面板显示当前文档的完整绝对路径和所在目录，并支持复制路径。
```

- [ ] **Step 2: Run the full test and static checks**

Run: `go test ./... -count=1`

Expected: all packages pass.

Run: `go vet ./...`

Expected: exit 0 and no output.

- [ ] **Step 3: Build**

Run: `printf '\n' | cmd.exe //c build.bat`

Expected: tests pass, version/build time injected, `OpenMD.exe` and `source.git.bundle` generated.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document file properties panel"
```
