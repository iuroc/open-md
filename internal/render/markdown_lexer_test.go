package render

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2/lexers"
)

func TestMarkdownSourceHighlightsTOMLFrontMatter(t *testing.T) {
	got := string(highlightSource([]byte("+++\ntitle = \"Hi\"\n+++\n\n# 标题\n")))
	for _, want := range []string{`class="o"`, `class="s2"`, `class="gh"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in TOML front matter source: %q", want, got)
		}
	}
	if strings.Contains(got, `class="err"`) {
		t.Fatalf("TOML front matter produced error tokens: %q", got)
	}
}

func TestMarkdownSourceHighlightsTables(t *testing.T) {
	got := string(highlightSource([]byte("| A | B |\n| --- | --- |\n| 1 | 2 |\n")))
	for _, want := range []string{`class="o"`, `class="k"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in table source: %q", want, got)
		}
	}
}

func TestMarkdownSourceHighlightsReferenceLinks(t *testing.T) {
	got := string(highlightSource([]byte("[ref]: https://example.com\n\n[文字][ref]\n")))
	for _, want := range []string{`class="nt"`, `class="na"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in reference link source: %q", want, got)
		}
	}
}

func TestMarkdownSourceHighlightsAutolinks(t *testing.T) {
	got := string(highlightSource([]byte("<https://example.com>\n")))
	if !strings.Contains(got, `class="na"`) {
		t.Fatalf("autolink missing URL token: %q", got)
	}
}

func TestMarkdownSourceHighlightsFootnotes(t *testing.T) {
	got := string(highlightSource([]byte("文字[^1]\n\n[^1]: 脚注\n")))
	if !strings.Contains(got, `class="ni"`) {
		t.Fatalf("footnote missing entity token: %q", got)
	}
}

func TestMarkdownSourceHighlightsSetextHeadings(t *testing.T) {
	got := string(highlightSource([]byte("标题\n====\n\n副标题\n----\n")))
	for _, want := range []string{`class="gh"`, `class="gu"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in setext source: %q", want, got)
		}
	}
}

func TestMarkdownSourceHighlightsHTMLTags(t *testing.T) {
	got := string(highlightSource([]byte("<span class=\"badge\">文字</span>\n")))
	for _, want := range []string{`class="nt"`, `class="na"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in HTML source: %q", want, got)
		}
	}
}

func TestMarkdownLexerMapsLessToCSS(t *testing.T) {
	lexer := lexers.Get("less")
	if lexer == nil {
		t.Fatal("less language has no lexer")
	}
	if lexer.Config().Name != "CSS" {
		t.Fatalf("less lexer = %q, want CSS", lexer.Config().Name)
	}
}

func TestMarkdownSourceDoesNotTreatHorizontalRuleAsSetext(t *testing.T) {
	got := string(highlightSource([]byte("普通文字\n\n---\n\n下一段\n")))
	if strings.Contains(got, `class="gh"`) || strings.Contains(got, `class="gu"`) {
		t.Fatalf("horizontal rule was treated as a setext heading: %q", got)
	}
}

func TestMarkdownSourceDoesNotTreatListBreakAsSetext(t *testing.T) {
	got := string(highlightSource([]byte("- 列表项\n---\n")))
	if strings.Contains(got, `class="gu"`) {
		t.Fatalf("list followed by rule was treated as setext: %q", got)
	}
}

func TestMarkdownSourceDoesNotReinterpretFencedCode(t *testing.T) {
	got := string(highlightSource([]byte("```text\n| A | B |\n标题\n====\n```\n")))
	if strings.Contains(got, `class="o"`) || strings.Contains(got, `class="gh"`) {
		t.Fatalf("fenced code body was reinterpreted as Markdown: %q", got)
	}
}

func TestMarkdownSourceDoesNotTreatIndentedCodeAsTable(t *testing.T) {
	got := string(highlightSource([]byte("    | A | B |\n")))
	if strings.Contains(got, `class="o"`) {
		t.Fatalf("indented code was treated as a table: %q", got)
	}
}

func TestMarkdownSourceRequiresClosingTOMLDelimiter(t *testing.T) {
	got := string(highlightSource([]byte("+++\ntitle = \"Hi\"\n")))
	if strings.Contains(got, `class="o"`) {
		t.Fatalf("unclosed block was treated as TOML front matter: %q", got)
	}
}
