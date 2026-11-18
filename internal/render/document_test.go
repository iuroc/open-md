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

func TestHeadingSlugEnglishCompatibility(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# Hello World\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc.HTML), `id="doc-0-hello-world"`) {
		t.Fatalf("English heading ID changed: %s", doc.HTML)
	}
}

func TestHeadingSlugChineseAndDuplicates(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# 中文标题：开始\n\n## 中文标题：开始\n\n# **粗体** [链接](other.md)\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	for _, want := range []string{`id="doc-0-中文标题开始"`, `id="doc-0-中文标题开始-2"`, `id="doc-0-粗体-链接"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in %s", want, html)
		}
	}
	if strings.Contains(html, "heading") {
		t.Fatalf("fallback heading ID still present: %s", html)
	}
}

func TestFootnoteIDsAreDocumentScoped(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# A\n\ntext[^1]\n\n[^1]: note\n"), `C:\docs\a.md`, RenderOptions{
		IDPrefix:     "doc-0",
		CurrentDocID: "doc-0",
	})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	for _, want := range []string{`id="doc-0-fn:1"`, `id="doc-0-fnref:1"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in %s", want, html)
		}
	}
}

func TestFrontMatterYAMLIsStripped(t *testing.T) {
	doc, err := RenderMarkdown([]byte("---\ntitle: Hi\ndate: 2026-01-01\n---\n\n# 正文标题\n\nBody\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if strings.Contains(html, "title: Hi") || strings.Contains(html, "<hr>") {
		t.Fatalf("front matter leaked into HTML: %s", html)
	}
	if doc.Title != "正文标题" {
		t.Fatalf("title = %q", doc.Title)
	}
	if len(doc.TOC) != 1 || doc.TOC[0].Text != "正文标题" {
		t.Fatalf("toc = %#v", doc.TOC)
	}
}

func TestFrontMatterTOMLIsStripped(t *testing.T) {
	doc, err := RenderMarkdown([]byte("+++\ntitle = \"Hi\"\n+++\n\n# 正文标题\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(doc.HTML), "title =") {
		t.Fatalf("TOML front matter leaked: %s", doc.HTML)
	}
	if doc.Title != "正文标题" {
		t.Fatalf("title = %q", doc.Title)
	}
	if len(doc.TOC) != 1 || doc.TOC[0].Text != "正文标题" {
		t.Fatalf("toc = %#v", doc.TOC)
	}
}

func TestFrontMatterWithoutClosingDelimiterIsContent(t *testing.T) {
	doc, err := RenderMarkdown([]byte("---\n\n# 标题\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc.HTML), "<hr>") {
		t.Fatalf("unclosed front matter should remain content: %s", doc.HTML)
	}
	if doc.Title != "标题" {
		t.Fatalf("title = %q", doc.Title)
	}
}

func TestFrontMatterHorizontalRuleAfterContentIsPreserved(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# 标题\n\n---\n\n内容\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc.HTML), "<hr>") {
		t.Fatalf("horizontal rule removed: %s", doc.HTML)
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
