package render

import (
	"strings"
	"testing"
)

func TestRenderMarkdownGFMAndTOC(t *testing.T) {
	src := []byte("# 标题\n\n## 用法\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n- [x] done\n")
	doc, err := RenderMarkdown(src, `C:\docs\README.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "标题" {
		t.Fatalf("title = %q", doc.Title)
	}
	if len(doc.TOC) != 2 {
		t.Fatalf("toc len = %d", len(doc.TOC))
	}
	if doc.TOC[0].ID == "" || doc.TOC[1].ID == "" {
		t.Fatal("missing heading IDs")
	}
	html := string(doc.HTML)
	if !strings.Contains(html, "<table>") {
		t.Fatal("GFM table missing")
	}
	if !strings.Contains(html, `type="checkbox"`) {
		t.Fatal("task list missing")
	}
}

func TestRenderMarkdownTOCOnlyIncludesLevelsOneToThree(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# One\n\n## Two\n\n### Three\n\n#### Four\n"), "note.md", RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.TOC) != 3 {
		t.Fatalf("toc len = %d, want 3", len(doc.TOC))
	}
	if doc.TOC[2].Level != 3 {
		t.Fatalf("last toc level = %d", doc.TOC[2].Level)
	}
}

func TestRenderMarkdownRawHTMLIsOmitted(t *testing.T) {
	doc, err := RenderMarkdown([]byte("<script>alert(1)</script>\n\nhello"), "note.md", RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(doc.HTML), "<script>") {
		t.Fatal("raw HTML was rendered")
	}
	if !strings.Contains(string(doc.HTML), "hello") {
		t.Fatal("normal text missing")
	}
}

func TestRenderMarkdownTitleFallbackAndHighlighting(t *testing.T) {
	doc, err := RenderMarkdown([]byte("```go\nfunc main() {}\n```\n"), `C:\docs\no-title.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "no-title" {
		t.Fatalf("title = %q", doc.Title)
	}
	if !strings.Contains(string(doc.HTML), "chroma") {
		t.Fatal("Chroma classes missing")
	}
}
