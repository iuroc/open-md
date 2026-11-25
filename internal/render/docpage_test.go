package render

import (
	"html"
	"html/template"
	"strings"
	"testing"
)

func TestRenderDocPageTOCIncludesSourceLine(t *testing.T) {
	out, err := RenderDocPage(DocPageData{
		Title:         "A",
		InlineSection: true,
		Doc: Document{
			ID:    "doc-0",
			Title: "A",
			HTML:  template.HTML("<p>a</p>"),
			TOC:   []Heading{{Level: 1, Text: "T", ID: "t", Line: 3}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `data-toc-line="3"`) {
		t.Fatal("doc page TOC is missing the source line")
	}
}

func TestRenderDocPageShowsAuthorIDInsteadOfLink(t *testing.T) {
	out, err := RenderDocPage(DocPageData{
		Title:      "A",
		Doc:        Document{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
		AuthorName: "内部作者",
		AuthorID:   "12345",
		AuthorURL:  "https://example.invalid/profile",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))
	if !strings.Contains(s, `name="author" content="内部作者（12345）"`) {
		t.Fatal("author meta must include the employee ID")
	}
	if !strings.Contains(s, "内部作者（12345）") {
		t.Fatal("author line must include the employee ID")
	}
	if strings.Contains(s, `href="https://example.invalid/profile"`) {
		t.Fatal("author link must not be rendered when an employee ID is set")
	}
}
