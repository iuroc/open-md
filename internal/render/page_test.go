package render

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	for _, want := range []string{"<!doctype html>", "Content-Security-Policy", "<style>", "<script>", `data-doc-id="doc-0"`, `class="toc-template"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(s, "<link rel=") || strings.Contains(s, "script src=") {
		t.Fatal("external asset referenced")
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
