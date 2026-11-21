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

func TestRenderPageSourceToggleAndBlocks(t *testing.T) {
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
		`id="source-toggle"`,
		`aria-pressed="false"`,
		`aria-label="查看源码"`,
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
