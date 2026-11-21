package render

import (
	"html"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openmd/internal/buildinfo"
)

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
		`class="copy-path panel-action"`,
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

func TestRenderPageInfoPanelUsesTabs(t *testing.T) {
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
	s := string(out)
	for _, want := range []string{
		`class="settings-title">OpenMD</span>`,
		`class="settings-tabs" role="tablist" aria-label="信息与设置分类"`,
		`id="settings-tab-button-file" data-settings-tab="file" role="tab" aria-selected="true" aria-controls="settings-panel-file" tabindex="0">文件</button>`,
		`id="settings-tab-button-settings" data-settings-tab="settings" role="tab" aria-selected="false" aria-controls="settings-panel-settings" tabindex="-1">设置</button>`,
		`id="settings-tab-button-about" data-settings-tab="about" role="tab" aria-selected="false" aria-controls="settings-panel-about" tabindex="-1">关于</button>`,
		`id="settings-panel-file" class="settings-tabpanel" role="tabpanel" aria-labelledby="settings-tab-button-file" data-settings-panel="file"`,
		`id="settings-panel-settings" class="settings-tabpanel" role="tabpanel" aria-labelledby="settings-tab-button-settings" data-settings-panel="settings" hidden`,
		`id="settings-panel-about" class="settings-tabpanel" role="tabpanel" aria-labelledby="settings-tab-button-about" data-settings-panel="about" hidden`,
		`id="file-path"`,
		`id="copy-path"`,
		`data-theme="blue"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in panel tabs", want)
		}
	}
	for _, banned := range []string{
		`class="settings-label">文件<`,
		`class="settings-label">设置<`,
		`class="settings-label">关于<`,
		`class="settings-label">关于 OpenMD<`,
	} {
		if strings.Contains(s, banned) {
			t.Fatalf("duplicated group heading %q", banned)
		}
	}
}

func TestBuildFilePageNormalizesRelativeArgument(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Title\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	page, err := BuildFilePage("note.md")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "note.md")
	if page.Documents[0].Path != want {
		t.Fatalf("path = %q, want %q", page.Documents[0].Path, want)
	}
	if page.Documents[0].Dir != dir {
		t.Fatalf("dir = %q, want %q", page.Documents[0].Dir, dir)
	}
}

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

func TestReadFileLimitedRejectsOverLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("1234567890"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := readFileLimited(path, 4); err == nil {
		t.Fatal("over-limit read succeeded")
	}
	got, err := readFileLimited(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "1234567890" {
		t.Fatalf("data = %q", got)
	}
}

func TestBuildFilePageRejectsOversizedFile(t *testing.T) {
	oldLimit := maxSingleFileBytes
	maxSingleFileBytes = 4
	defer func() { maxSingleFileBytes = oldLimit }()

	dir := t.TempDir()
	path := filepath.Join(dir, "big.md")
	if err := os.WriteFile(path, []byte("# 1234567890\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := BuildFilePage(path)
	if err == nil {
		t.Fatal("oversized file was accepted")
	}
	if !strings.Contains(err.Error(), "big.md") || !strings.Contains(err.Error(), "上限") {
		t.Fatalf("error does not identify the file and limit: %v", err)
	}
}

func TestBuildFilePage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("# Hello\n\nBody\n"), 0644); err != nil {
		t.Fatal(err)
	}
	page, err := BuildFilePage(path)
	if err != nil {
		t.Fatal(err)
	}
	if page.Mode != "file" {
		t.Fatalf("mode = %q", page.Mode)
	}
	if page.Title != "Hello" {
		t.Fatalf("title = %q", page.Title)
	}
	if len(page.Documents) != 1 || page.Active != 0 {
		t.Fatalf("docs = %#v active = %d", page.Documents, page.Active)
	}
	if page.Documents[0].ID != "doc-0" {
		t.Fatalf("doc ID = %q", page.Documents[0].ID)
	}
	if !strings.Contains(string(page.Documents[0].HTML), `id="doc-0-hello"`) {
		t.Fatalf("heading ID not prefixed: %s", page.Documents[0].HTML)
	}
}

func TestRenderPageContainsEmbeddedUIAndCSP(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "T",
		Documents: []Document{{
			ID:    "doc-0",
			Title: "T",
			HTML:  template.HTML("<p>hi</p>"),
			TOC:   []Heading{{Level: 1, Text: "T", ID: "doc-0-t"}},
		}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{"<!doctype html>", "Content-Security-Policy", "<style>", "<script>", `data-doc-id="doc-0"`, `class="toc-template"`, `id="menu-toggle"`, `id="width-toggle"`, `id="settings-button"`, `id="settings-panel"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if !strings.Contains(s, `rel="icon"`) || !strings.Contains(s, "data:image/png;base64,") {
		t.Fatal("embedded favicon missing")
	}
	if strings.Contains(s, `rel="stylesheet"`) || strings.Contains(s, "script src=") {
		t.Fatal("external asset referenced")
	}
}

func TestRenderPageViewModeMenuAndSourceBlocks(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "A",
		Documents: []Document{
			{
				ID:         "doc-0",
				Title:      "A",
				RelPath:    "a.md",
				HTML:       template.HTML("<p>a</p>"),
				SourceHTML: template.HTML(`<span class="gh"># A</span>`),
				HasSource:  true,
			},
			{
				ID:      "doc-1",
				Title:   "B",
				RelPath: "b.md",
				HTML:    template.HTML("<p>b</p>"),
			},
		},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		`id="mode-button"`,
		`id="mode-menu"`,
		`role="menu"`,
		`role="menuitemradio"`,
		`data-mode="render"`,
		`data-mode="source"`,
		`data-mode="split"`,
		`aria-checked="true"`,
		`aria-label="阅读模式：阅读"`,
		`阅读模式`,
		`只看排版后的正文`,
		`整页显示 Markdown 原文`,
		`双视图`,
		`左侧源码，右侧正文，同步滚动`,
		`class="doc-rendered"`,
		`class="source-view"`,
		`class="source-toolbar"`,
		`class="source-filename"`,
		`class="chroma light source-code"`,
		`data-has-source="true"`,
		`data-has-source="false"`,
		`复制源码`,
		`该文档没有可显示的源码`,
		`<span class="gh"># A</span>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in page", want)
		}
	}
	if got := strings.Count(s, `class="source-copy panel-action"`); got != 1 {
		t.Fatalf("source copy button count = %d, want 1", got)
	}
}

func TestRenderPageDirectoryUsesSidebarTabs(t *testing.T) {
	page := &Page{
		Mode:  "dir",
		Title: "docs",
		Documents: []Document{
			{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
			{ID: "doc-1", Title: "B", HTML: template.HTML("<p>b</p>")},
		},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		`role="tablist"`,
		`data-tab="toc"`,
		`data-tab="files"`,
		`id="doc-list"`,
		`id="toc"`,
		`data-tab-panel="toc"`,
		`data-tab-panel="files"`,
		`class="sidebar-tab active" id="tab-button-files"`,
		`aria-selected="false" aria-controls="tab-panel-toc"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(s, `data-tab-panel="files" role="tabpanel" aria-labelledby="tab-button-files" hidden`) {
		t.Fatal("file list panel must be visible by default")
	}
	if !strings.Contains(s, `data-tab-panel="toc" role="tabpanel" aria-labelledby="tab-button-toc" hidden`) {
		t.Fatal("outline panel must be hidden by default")
	}
}

func TestRenderPageDirectoryRendersFileTree(t *testing.T) {
	docs := []Document{
		{ID: "doc-0", Title: "Root", RelPath: "README.md", HTML: template.HTML("<p>root</p>")},
		{ID: "doc-1", Title: "Spec Delta", RelPath: "specs/source-view/spec.md", HTML: template.HTML("<p>spec</p>")},
	}
	page := &Page{
		Mode:      "dir",
		Title:     "docs",
		Active:    1,
		Documents: docs,
	}
	page.Tree = buildTree(docs, "doc-1")
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		`class="doc-tree"`,
		`class="tree-root"`,
		`data-tree-folder`,
		`aria-expanded="true"`,
		`aria-expanded="false"`,
		`class="tree-label">specs</span>`,
		`class="tree-label">spec.md</span>`,
		`title="Spec Delta — specs/source-view/spec.md"`,
		`class="tree-file-button active"`,
		`aria-current="true"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in tree markup", want)
		}
	}
	if strings.Contains(s, `class="tree-label">Spec Delta</span>`) {
		t.Fatal("file node must use basename as its label")
	}
}

func TestReaderCSSRendersFileTree(t *testing.T) {
	if !strings.Contains(readerCSS, "padding-left: 21px;") {
		t.Fatal("tree file rule missing left padding")
	}
	for _, want := range []string{"text-overflow: ellipsis;", "white-space: nowrap;"} {
		if !strings.Contains(readerCSS, want) {
			t.Fatalf("tree label rule missing %q", want)
		}
	}
	hidden := readerCSSRule(t, ".tree-children[hidden] {")
	if !strings.Contains(hidden, "display: none") {
		t.Fatalf("collapsed tree children must be hidden: %s", hidden)
	}
	if !strings.Contains(readerCSS, ".tree-file-link.active") || !strings.Contains(readerCSS, "background: var(--accent-soft);") {
		t.Fatal("active tree file must use the soft accent background")
	}
}

func TestReaderJSSupportsSoftDocumentNavigation(t *testing.T) {
	for _, want := range []string{
		"isDocPageLink",
		"window.OPENMD_PAGES",
		"sectionHTML",
		"ownFile",
		`querySelector("#doc-content .doc-section")`,
		"revealPage",
		"openmd-ready",
		"openmd-hash-load",
		"loadSoftPage",
		"navGeneration",
		"hashchange",
		"parseDocHash",
	} {
		if !strings.Contains(readerJS, want) {
			t.Fatalf("reader.js missing %q", want)
		}
	}
	if strings.Contains(readerJS, "DOMParser") {
		t.Fatal("payloads must not be parsed as full documents")
	}
}

func TestDocPageHidesUntilReady(t *testing.T) {
	for _, want := range []string{
		"html.openmd-booting body { visibility: hidden; }",
		"html.openmd-booting.openmd-ready body { visibility: visible; }",
		"openmd-booting",
		"openmd-ready",
		`history.scrollRestoration = "manual"`,
	} {
		if !strings.Contains(docPageTemplateSource, want) {
			t.Fatalf("doc page missing %q", want)
		}
	}
	hidden := readerCSSRule(t, "html.openmd-hash-load #doc-content {")
	if !strings.Contains(hidden, "visibility: hidden;") {
		t.Fatalf("hash-loaded content must start hidden: %s", hidden)
	}
}

func TestReaderJSRestoresViewModeOnReload(t *testing.T) {
	for _, want := range []string{
		"modeStorageKey",
		"legacySourceKey",
		"storedMode",
		"restoreViewMode",
		"sessionStorage.getItem(modeStorageKey)",
	} {
		if !strings.Contains(readerJS, want) {
			t.Fatalf("reader.js missing %q", want)
		}
	}
	if strings.Contains(readerJS, `navigationType() === "reload"`) {
		t.Fatal("reload must not clear the view mode")
	}
}

func TestReaderJSSupportsSplitView(t *testing.T) {
	for _, want := range []string{
		`"mode-button"`,
		`"mode-menu"`,
		".mode-option",
		"applyMode",
		"resolveMode",
		"currentScrollRatio",
		"applyScrollRatio",
		"syncPanes",
		"renderedPane",
		"sourcePane",
		`section.querySelector(".source-view")`,
		`target.classList.contains("source-view")`,
		"elementRatio",
		"split-mode",
		"setElementTop",
	} {
		if !strings.Contains(readerJS, want) {
			t.Fatalf("reader.js missing %q", want)
		}
	}
}

func TestReaderCSSSupportsSplitView(t *testing.T) {
	section := readerCSSRule(t, "body.split-mode .doc-section.markdown-body {")
	for _, want := range []string{
		"display: grid;",
		"grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);",
		"height: 100%;",
		"max-width: none;",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("split section rule missing %q: %s", want, section)
		}
	}
	source := readerCSSRule(t, "body.split-mode .doc-section .source-view {")
	for _, want := range []string{
		"display: block;",
		"grid-column: 1;",
		"grid-row: 1;",
		"overflow-y: auto;",
		"border-right: 1px solid var(--border);",
		"scrollbar-width: none;",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("split source rule missing %q: %s", want, source)
		}
	}
	scrollbar := readerCSSRule(t, "body.split-mode .doc-section .source-view::-webkit-scrollbar {")
	if !strings.Contains(scrollbar, "display: none;") {
		t.Fatalf("split source scrollbar must be hidden: %s", scrollbar)
	}
	for _, banned := range []string{"background: var(--bg-subtle);", "border-bottom"} {
		if strings.Contains(source, banned) {
			t.Fatalf("split source pane must not look like a shell, found %q: %s", banned, source)
		}
	}
	if strings.Contains(readerCSS, "body.split-mode .doc-section .source-toolbar {") {
		t.Fatal("split source toolbar must keep the standalone source view styling")
	}
	if strings.Contains(readerCSS, "body.split-mode .doc-section .source-code.chroma {") {
		t.Fatal("split source code must keep the standalone source view typography")
	}
	rendered := readerCSSRule(t, "body.split-mode .doc-section .doc-rendered {")
	for _, want := range []string{
		"grid-column: 2;",
		"grid-row: 1;",
		"overflow-y: auto;",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("split rendered rule missing %q: %s", want, rendered)
		}
	}
	if !strings.Contains(readerCSS, "body.split-mode #doc-content {") ||
		!strings.Contains(readerCSS, "height: calc(100vh - var(--chrome-height));") {
		t.Fatal("split mode must fill the viewport below the topbar")
	}
}

func TestReaderCSSPrintFollowsViewMode(t *testing.T) {
	printCSS := readerPrintCSS(t)
	for _, want := range []string{
		".app,",
		"min-height: 0;",
		"body.split-mode {",
		"overflow: visible;",
	} {
		if !strings.Contains(printCSS, want) {
			t.Fatalf("print block missing %q", want)
		}
	}
	content := cssRule(t, printCSS, "body.split-mode #doc-content {")
	for _, want := range []string{"height: auto;", "overflow: visible;"} {
		if !strings.Contains(content, want) {
			t.Fatalf("split print content rule missing %q: %s", want, content)
		}
	}
	section := cssRule(t, printCSS, "body.split-mode .doc-section.markdown-body {")
	for _, want := range []string{"display: block;", "height: auto;"} {
		if !strings.Contains(section, want) {
			t.Fatalf("split print section rule missing %q: %s", want, section)
		}
	}
	source := cssRule(t, printCSS, "body.split-mode .doc-section .source-view {")
	if !strings.Contains(source, "display: none;") {
		t.Fatalf("split print must hide the source pane: %s", source)
	}
	rendered := cssRule(t, printCSS, "body.split-mode .doc-section .doc-rendered {")
	for _, want := range []string{"overflow: visible;", "padding: 0;"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("split print rendered rule missing %q: %s", want, rendered)
		}
	}
}

func TestReaderCSSPrintWrapsCodeAndTables(t *testing.T) {
	printCSS := readerPrintCSS(t)
	pre := cssRule(t, printCSS, ".markdown-body pre,")
	for _, want := range []string{
		"overflow: visible !important;",
		"white-space: pre-wrap !important;",
		"overflow-wrap: anywhere !important;",
		"word-break: break-word !important;",
	} {
		if !strings.Contains(pre, want) {
			t.Fatalf("printed code rule missing %q: %s", want, pre)
		}
	}
	table := cssRule(t, printCSS, ".markdown-body table {")
	for _, want := range []string{"display: table;", "width: 100%;", "overflow: visible;"} {
		if !strings.Contains(table, want) {
			t.Fatalf("printed table rule missing %q: %s", want, table)
		}
	}
	cells := cssRule(t, printCSS, ".markdown-body table th,")
	for _, want := range []string{"white-space: normal;", "overflow-wrap: anywhere;"} {
		if !strings.Contains(cells, want) {
			t.Fatalf("printed table cell rule missing %q: %s", want, cells)
		}
	}
	source := cssRule(t, printCSS, ".doc-section .source-code.chroma {")
	for _, want := range []string{"white-space: pre-wrap !important;", "overflow-x: hidden !important;"} {
		if !strings.Contains(source, want) {
			t.Fatalf("printed source rule missing %q: %s", want, source)
		}
	}
}

func TestReaderCSSPrintUsesWhiteBackground(t *testing.T) {
	printCSS := readerPrintCSS(t)
	root := cssRule(t, printCSS, ":root {")
	for _, want := range []string{
		"--bg: #ffffff !important;",
		"--fg: #1f2328 !important;",
		"--border: #d1d9e0 !important;",
	} {
		if !strings.Contains(root, want) {
			t.Fatalf("print root rule missing %q: %s", want, root)
		}
	}
	for _, want := range []string{
		"background-color: #ffffff !important;",
		"color-scheme: light !important;",
		"--bgColor-default: #ffffff !important;",
		"--bgColor-muted: #f6f8fa !important;",
		"--bgColor-neutral-muted: #818b981f !important;",
		"--fgColor-default: #1f2328 !important;",
	} {
		if !strings.Contains(printCSS, want) {
			t.Fatalf("print block missing white background rule %q", want)
		}
	}
	if strings.Contains(printCSS, "background-color: transparent !important;") {
		t.Fatal("print must keep content backgrounds instead of forcing them transparent")
	}
	if strings.Contains(printCSS, "--bgColor-muted: #ffffff !important;") {
		t.Fatal("print must not flatten content backgrounds to pure white")
	}
	toolbar := cssRule(t, printCSS, ".source-toolbar {")
	if !strings.Contains(toolbar, "display: none !important;") {
		t.Fatalf("print must hide the source toolbar: %s", toolbar)
	}
	darkIndex := strings.Index(readerCSS, "@media print and (prefers-color-scheme: dark)")
	if darkIndex < 0 {
		t.Fatal("dark-mode print readability rule missing")
	}
	dark := readerCSS[darkIndex:]
	if !strings.Contains(dark, ".doc-section .chroma.dark span") ||
		!strings.Contains(dark, "color: #1f2328 !important;") {
		t.Fatalf("dark-mode print must keep code readable on white: %s", dark)
	}
}

func TestReaderCSSPrintOverridesStayScoped(t *testing.T) {
	printStart := strings.Index(readerCSS, "@media print {")
	if printStart < 0 {
		t.Fatal("print media block missing")
	}
	screenCSS := readerCSS[:printStart]
	for _, banned := range []string{
		"--bg: #ffffff !important;",
		"--bgColor-default: #ffffff !important;",
		"--fgColor-default: #1f2328 !important;",
		"color-scheme: light !important;",
	} {
		if strings.Contains(screenCSS, banned) {
			t.Fatalf("print-only override leaked into screen styles: %q", banned)
		}
	}
	for _, want := range []string{"beforeprint", "afterprint"} {
		if strings.Contains(readerJS, want) {
			t.Fatalf("print must not mutate runtime state via %q", want)
		}
	}
}

func TestReaderCSSHidesWidthToggleInSplit(t *testing.T) {
	rule := readerCSSRule(t, "body.split-mode #width-toggle {")
	if !strings.Contains(rule, "display: none;") {
		t.Fatalf("split mode must hide the width toggle: %s", rule)
	}
}

func TestSourceViewAlwaysWraps(t *testing.T) {
	for _, banned := range []string{
		"source-wrap-toggle",
		"source-no-wrap",
		"openmd-source-wrap",
		"setting-switch",
	} {
		for name, source := range map[string]string{
			"docpage.html": docPageTemplateSource,
			"page.html":    pageTemplateSource,
			"reader.js":    readerJS,
			"reader.css":   readerCSS,
		} {
			if strings.Contains(source, banned) {
				t.Fatalf("%s still contains removed source wrap feature %q", name, banned)
			}
		}
	}
	base := readerCSSRule(t, ".doc-section .source-code.chroma {")
	for _, want := range []string{
		"white-space: pre-wrap;",
		"overflow-wrap: anywhere;",
		"word-break: break-all;",
		"overflow-x: hidden;",
	} {
		if !strings.Contains(base, want) {
			t.Fatalf("source code rule missing always-wrap property %q: %s", want, base)
		}
	}
}

func TestReaderJSPersistsScrollPositions(t *testing.T) {
	for _, want := range []string{
		"openmd-scroll:",
		"saveBodyScroll",
		"restoreBodyScroll",
		"finishInitialReveal",
		"suppressTreeScrollOnce",
		"var restored = restoreSidebarScroll();",
		"scrollWindowTo",
		"currentScrollRatio",
		"applyScrollRatio",
		"markClickHash",
		"consumeClickHash",
		"preserveScroll",
		"restoreScroll",
	} {
		if !strings.Contains(readerJS, want) {
			t.Fatalf("reader.js missing %q", want)
		}
	}
}

func TestReaderJSClosesMobileSidebarOnNavigation(t *testing.T) {
	tocStart := strings.Index(readerJS, `toc.addEventListener("click"`)
	if tocStart < 0 {
		t.Fatal("toc click handler missing")
	}
	tocBlock := readerJS[tocStart:]
	if end := strings.Index(tocBlock, "});"); end >= 0 {
		tocBlock = tocBlock[:end]
	}
	if !strings.Contains(tocBlock, "closeMobileSidebar();") {
		t.Fatal("clicking an outline item must close the mobile sidebar")
	}
	treeStart := strings.Index(readerJS, `link.closest("#doc-list")`)
	if treeStart < 0 {
		t.Fatal("document tree navigation branch missing")
	}
	treeBlock := readerJS[treeStart:]
	if end := strings.Index(treeBlock, "}"); end >= 0 {
		treeBlock = treeBlock[:end]
	}
	if !strings.Contains(treeBlock, "closeMobileSidebar();") {
		t.Fatal("clicking a tree document must close the mobile sidebar")
	}
}

func TestReaderCSSDisablesTreeLinkDragAndUnderline(t *testing.T) {
	for _, want := range []string{
		"#doc-list a:hover",
		"-webkit-user-drag: none;",
	} {
		if !strings.Contains(readerCSS, want) {
			t.Fatalf("reader.css missing %q", want)
		}
	}
}

func TestReaderJSWiresFileTree(t *testing.T) {
	for _, want := range []string{
		`document.querySelectorAll("[data-doc-target]")`,
		`document.querySelectorAll("[data-tree-folder]")`,
		"revealTreeTarget",
		`aria-expanded`,
		`.tree-children`,
		`"dragstart"`,
		`#doc-list a[href]`,
	} {
		if !strings.Contains(readerJS, want) {
			t.Fatalf("reader.js missing %q", want)
		}
	}
}

func TestReaderCSSSourceHasNoOuterCard(t *testing.T) {
	rule := readerCSSRule(t, ".doc-section .source-code.chroma {")
	for _, want := range []string{
		"padding: 0;",
		"border: 0;",
		"border-radius: 0;",
		"background-color: transparent;",
		"color: var(--fg);",
	} {
		if !strings.Contains(rule, want) {
			t.Fatalf("source code rule missing %q: %s", want, rule)
		}
	}
	toolbar := readerCSSRule(t, ".source-toolbar {")
	for _, banned := range []string{"border", "background", "position: sticky"} {
		if strings.Contains(toolbar, banned) {
			t.Fatalf("source toolbar must stay borderless and static, found %q: %s", banned, toolbar)
		}
	}
}

func readerCSSRule(t *testing.T, selector string) string {
	t.Helper()
	return cssRule(t, readerCSS, selector)
}

func cssRule(t *testing.T, css, selector string) string {
	t.Helper()
	start := strings.Index(css, selector)
	if start < 0 {
		t.Fatalf("CSS rule %q missing", selector)
	}
	end := strings.Index(css[start:], "}")
	if end < 0 {
		t.Fatalf("CSS rule %q has no closing brace", selector)
	}
	return css[start : start+end]
}

func readerPrintCSS(t *testing.T) string {
	t.Helper()
	start := strings.Index(readerCSS, "@media print {")
	if start < 0 {
		t.Fatal("print media block missing")
	}
	return readerCSS[start:]
}

func TestReaderCSSSourceUsesBodyFontSize(t *testing.T) {
	start := strings.Index(readerCSS, ".doc-section .source-code.chroma {")
	if start < 0 {
		t.Fatal("source code rule missing")
	}
	end := strings.Index(readerCSS[start:], "}")
	if end < 0 {
		t.Fatal("source code rule has no closing brace")
	}
	rule := readerCSS[start : start+end]
	for _, want := range []string{"font-size: 1rem;", "line-height: 1.5;"} {
		if !strings.Contains(rule, want) {
			t.Fatalf("source code rule missing %q: %s", want, rule)
		}
	}
}

func TestReaderCSSColorsMarkdownTokens(t *testing.T) {
	for _, selector := range []string{
		".doc-section .chroma.light .gh,",
		".doc-section .chroma.light .gu,",
		".doc-section .chroma.light .gs,",
		".doc-section .chroma.light .ge,",
		".doc-section .chroma.light .na,",
	} {
		if !strings.Contains(readerCSS, selector) {
			t.Fatalf("missing Markdown token selector %q", selector)
		}
	}
	heading := readerCSSRule(t, ".doc-section .chroma.light .gh,")
	for _, want := range []string{"color: var(--accent);", "font-weight: 700;"} {
		if !strings.Contains(heading, want) {
			t.Fatalf("heading rule missing %q: %s", want, heading)
		}
	}
	emphasis := readerCSSRule(t, ".doc-section .chroma.light .ge,")
	if !strings.Contains(emphasis, "font-style: italic;") {
		t.Fatalf("emphasis rule missing italic: %s", emphasis)
	}
}

func TestReaderCSSOffsetsFootnoteAnchors(t *testing.T) {
	rule := readerCSSRule(t, `.doc-section sup[id*="fnref:"],`)
	if !strings.Contains(rule, ".doc-section .footnotes li") {
		t.Fatalf("footnote definition target missing from rule: %s", rule)
	}
	if !strings.Contains(rule, "scroll-margin-top: 80px;") {
		t.Fatalf("footnote anchors do not account for the topbar: %s", rule)
	}
}

func TestReaderCSSKeepsTableHeadersOnOneLine(t *testing.T) {
	if !strings.Contains(readerCSS, ".markdown-body table th {") {
		t.Fatal("table header rule missing")
	}
	if !strings.Contains(readerCSS, "white-space: nowrap;") {
		t.Fatal("table header rule must keep headers on one line")
	}
}

func TestRenderPageIncludesGithubMarkdownCSS(t *testing.T) {
	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `class="doc-section markdown-body"`) {
		t.Fatal("markdown body class missing")
	}
	if !strings.Contains(s, `--fgColor-default`) {
		t.Fatal("github-markdown-css not embedded")
	}
}

func TestRenderPageFileModeHasNoFileListTab(t *testing.T) {
	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, `data-tab="files"`) {
		t.Fatal("single-file page must not render a file-list tab")
	}
	if !strings.Contains(s, `id="toc"`) {
		t.Fatal("single-file page must render the TOC")
	}
}

func TestRenderPageFileModeWithLinkedDocsDefaultsToOutlineTab(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "A",
		Documents: []Document{
			{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
			{ID: "doc-1", Title: "B", RelPath: `..\other\b.md`, Linked: true, HTML: template.HTML("<p>b</p>")},
		},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "链接的文档") {
		t.Fatal("linked documents must not introduce a second heading")
	}
	if !strings.Contains(s, `class="sidebar-tab active" id="tab-button-toc"`) {
		t.Fatal("outline tab should be active by default")
	}
	if !strings.Contains(s, `data-tab="files"`) {
		t.Fatal("document tab missing")
	}
	if !strings.Contains(s, `data-tab-panel="files" role="tabpanel" aria-labelledby="tab-button-files" hidden`) {
		t.Fatal("document panel should be hidden by default")
	}
	if !strings.Contains(s, `data-doc-target="doc-0"`) || !strings.Contains(s, `data-doc-target="doc-1"`) {
		t.Fatal("document list must contain both documents")
	}
	if !strings.Contains(s, `class="doc-item linked"`) {
		t.Fatal("linked document marker missing")
	}
}

func TestHighlightCSSIncludesBothModes(t *testing.T) {
	light, dark := HighlightCSS()
	if !strings.Contains(light, ".chroma.light") {
		t.Fatal("light Chroma CSS missing mode class")
	}
	if !strings.Contains(dark, ".chroma.dark") {
		t.Fatal("dark Chroma CSS missing mode class")
	}
}

func TestRenderPageEscapesTitle(t *testing.T) {
	page := &Page{
		Mode:      "file",
		Title:     `<script>alert(1)</script>`,
		Documents: []Document{{ID: "doc-0", Title: "x"}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "<script>alert(1)</script>") {
		t.Fatal("title was not escaped")
	}
}

func TestRenderPageShowsBuildInfoAndReleaseLink(t *testing.T) {
	oldVersion, oldBuildTime := buildinfo.Version, buildinfo.BuildTime
	defer func() {
		buildinfo.Version, buildinfo.BuildTime = oldVersion, oldBuildTime
	}()
	buildinfo.Version = "2026.11.18-0959"
	buildinfo.BuildTime = "2026-11-18T09:59:12+08:00"

	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))
	for _, want := range []string{
		`id="settings-panel"`,
		`name="author" content="github.com/iuroc"`,
		`href="https://github.com/iuroc"`,
		"github.com/iuroc",
		`data-theme="blue"`,
		`data-theme="teal"`,
		"v2026.11.18-0959",
		"2026-11-18 09:59:12 +08:00",
		"https://github.com/iuroc/open-md/releases",
		`class="release-link panel-action"`,
		"晚于当前构建时间",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestRenderPageHidesReleaseLinkForDevBuild(t *testing.T) {
	oldVersion, oldBuildTime := buildinfo.Version, buildinfo.BuildTime
	defer func() {
		buildinfo.Version, buildinfo.BuildTime = oldVersion, oldBuildTime
	}()
	buildinfo.Version = "dev"
	buildinfo.BuildTime = ""

	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "github.com/iuroc/open-md") {
		t.Fatal("dev build must not expose the release URL")
	}
	if !strings.Contains(s, "开发版") {
		t.Fatal("dev build label missing")
	}
}
