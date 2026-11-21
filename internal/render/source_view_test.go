package render

import (
	"errors"
	"html"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func stripHTMLTags(s string) string {
	return html.UnescapeString(htmlTagPattern.ReplaceAllString(s, ""))
}

type failingLexer struct{}

func (failingLexer) Config() *chroma.Config { return &chroma.Config{Name: "failing"} }

func (failingLexer) Tokenise(*chroma.TokeniseOptions, string) (chroma.Iterator, error) {
	return nil, errors.New("tokenise failed")
}

func (failingLexer) SetRegistry(*chroma.LexerRegistry) chroma.Lexer { return failingLexer{} }

func (failingLexer) SetAnalyser(func(string) float32) chroma.Lexer { return failingLexer{} }

func (failingLexer) AnalyseText(string) float32 { return 0 }

func TestHighlightSourceProducesTokenHTML(t *testing.T) {
	got := string(highlightSource([]byte("# 标题\n\n正文 **粗体**\n")))
	if !strings.Contains(got, `<span class="`) {
		t.Fatalf("source tokens are not highlighted: %q", got)
	}
	if strings.Contains(got, "<pre") || strings.Contains(got, "<code") {
		t.Fatalf("source HTML must not include a wrapper: %q", got)
	}
	for _, want := range []string{"标题", "正文", "粗体"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestHighlightSourceHighlightsIndentedFence(t *testing.T) {
	source := []byte("- item\n\n  ```go\n  func main() {}\n  ```\n")
	got := string(highlightSource(source))
	if !strings.Contains(got, `class="kd"`) && !strings.Contains(got, `class="nf"`) {
		t.Fatalf("indented Go fence is not highlighted: %q", got)
	}
	if plain := stripHTMLTags(got); plain != string(source) {
		t.Fatalf("source text changed:\n got %q\nwant %q", plain, source)
	}
}

func TestHighlightSourceTokenClassesForMarkdownSyntax(t *testing.T) {
	got := string(highlightSource([]byte("# 标题\n\n**粗体** 和 *斜体* 和 [链接](https://example.com)\n")))
	for _, want := range []string{`class="gh"`, `class="gs"`, `class="ge"`, `class="na"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestHighlightSourceEmptyInput(t *testing.T) {
	if got := string(highlightSource(nil)); got != "" {
		t.Fatalf("empty source = %q, want empty", got)
	}
}

func TestHighlightSourceFallsBackToEscapedText(t *testing.T) {
	got := string(highlightSourceWith(failingLexer{}, []byte("<script>alert(1)</script>")))
	if !strings.Contains(got, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("fallback did not escape source: %q", got)
	}
	if strings.Contains(got, "<script>") {
		t.Fatalf("fallback emitted raw HTML: %q", got)
	}
}

func TestBuildPageSourceForReadableDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.md")
	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    path,
		RelPath: "note.md",
		Source:  []byte("# 标题\n\n正文\n"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	doc := page.Documents[0]
	if !doc.HasSource {
		t.Fatal("readable document has no source")
	}
	if !strings.Contains(string(doc.SourceHTML), `<span class="`) {
		t.Fatalf("source HTML is not highlighted: %q", doc.SourceHTML)
	}
	if !strings.Contains(string(doc.SourceHTML), "标题") {
		t.Fatalf("source content missing: %q", doc.SourceHTML)
	}
}

func TestBuildPageSourceForEmptyDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.md")
	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    path,
		RelPath: "empty.md",
		Source:  nil,
	}})
	if err != nil {
		t.Fatal(err)
	}
	doc := page.Documents[0]
	if !doc.HasSource {
		t.Fatal("empty but readable document must still have a source")
	}
	if doc.SourceHTML != "" {
		t.Fatalf("empty source HTML = %q", doc.SourceHTML)
	}
}

func TestBuildPageSourceForReadFailure(t *testing.T) {
	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    filepath.Join(t.TempDir(), "missing.md"),
		RelPath: "missing.md",
		ReadErr: errors.New("无法读取"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	doc := page.Documents[0]
	if doc.HasSource {
		t.Fatal("failed document must not advertise a source")
	}
	if doc.SourceHTML != "" {
		t.Fatalf("failed source HTML = %q", doc.SourceHTML)
	}
}

func TestSourceKeepsFrontMatterWhileRenderedHTMLStripsIt(t *testing.T) {
	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    filepath.Join(t.TempDir(), "note.md"),
		RelPath: "note.md",
		Source:  []byte("---\ntitle: Hi\ndate: 2026-01-01\n---\n\n# 正文标题\n"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	doc := page.Documents[0]
	source := string(doc.SourceHTML)
	for _, want := range []string{"title", "Hi", "2026-01-01"} {
		if !strings.Contains(source, want) {
			t.Fatalf("source view missing front matter %q: %q", want, source)
		}
	}
	rendered := string(doc.HTML)
	for _, banned := range []string{"title: Hi", "2026-01-01", "<hr>"} {
		if strings.Contains(rendered, banned) {
			t.Fatalf("rendered HTML leaked front matter %q: %q", banned, rendered)
		}
	}
}

func TestBuildPageSourceFallsBackWhenHighlightingFails(t *testing.T) {
	oldHighlighter := sourceHighlighter
	sourceHighlighter = func(source []byte) template.HTML {
		return highlightSourceWith(failingLexer{}, source)
	}
	defer func() { sourceHighlighter = oldHighlighter }()

	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    filepath.Join(t.TempDir(), "note.md"),
		RelPath: "note.md",
		Source:  []byte("<b>raw</b>\n"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	doc := page.Documents[0]
	if !doc.HasSource {
		t.Fatal("document with fallback source must keep HasSource")
	}
	if !strings.Contains(string(doc.SourceHTML), "&lt;b&gt;raw&lt;/b&gt;") {
		t.Fatalf("source fallback not escaped: %q", doc.SourceHTML)
	}
}
