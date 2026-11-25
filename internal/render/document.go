package render

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"html/template"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Heading struct {
	Level int
	Text  string
	ID    string
	// Line is the zero-based line in the original source file, including any
	// BOM and front matter that are stripped before Markdown parsing.
	Line int
}

type Document struct {
	Path       string
	Dir        string
	RelPath    string
	ID         string
	Title      string
	HTML       template.HTML
	SourceHTML template.HTML
	HasSource  bool
	TOC        []Heading
	Linked     bool
	Active     bool
}

type RenderOptions struct {
	IDPrefix     string
	CurrentDocID string
	KnownDocs    map[string]string
	PageLinks    bool
}

func newMarkdown(docID string) goldmark.Markdown {
	var footnote goldmark.Extender = extension.Footnote
	if docID != "" {
		footnote = extension.NewFootnote(extension.WithFootnoteIDPrefix(docID + "-"))
	}
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithModeClasses(true),
				),
			),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(util.Prioritized(&safeLinkRenderer{}, 100)),
		),
	)
}

// highlightSource renders the original Markdown as Chroma token HTML. The
// template owns the surrounding <pre>, so the output only contains token
// spans and escaped text.
func highlightSource(source []byte) template.HTML {
	if len(source) == 0 {
		return template.HTML("")
	}
	lexer := lexers.Get("markdown")
	if lexer != nil {
		lexer = chroma.Coalesce(lexer)
	}
	if lexer == nil {
		return escapedSource(source)
	}
	normalized, indents := normalizeIndentedFences(source)
	highlighted, ok := highlightTokens(lexer, []byte(normalized))
	if !ok {
		return escapedSource(source)
	}
	if indents == nil {
		return highlighted
	}
	return reindentSourceHTML(highlighted, indents)
}

// sourceHighlighter is replaceable in tests to exercise the fallback path.
var sourceHighlighter = highlightSource

func highlightSourceWith(lexer chroma.Lexer, source []byte) template.HTML {
	if len(source) == 0 {
		return template.HTML("")
	}
	if lexer == nil {
		return escapedSource(source)
	}
	highlighted, ok := highlightTokens(lexer, source)
	if !ok {
		return escapedSource(source)
	}
	return highlighted
}

func highlightTokens(lexer chroma.Lexer, source []byte) (template.HTML, bool) {
	iterator, err := lexer.Tokenise(nil, string(source))
	if err != nil || iterator == nil {
		return "", false
	}
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.WithModeClasses(true),
		chromahtml.PreventSurroundingPre(true),
	)
	var buf bytes.Buffer
	if err := formatter.Format(&buf, styles.Get("github"), iterator); err != nil {
		return "", false
	}
	return template.HTML(buf.String()), true
}

func escapedSource(source []byte) template.HTML {
	return template.HTML(stdhtml.EscapeString(string(source)))
}

// normalizeIndentedFences moves fenced code blocks that start with leading
// whitespace (common inside list items) to column zero before tokenising.
// Chroma's Markdown lexer only delegates fences at column zero; without this,
// indented fences are treated as inline backtick strings and lose language
// highlighting. The returned slice records the removed prefix per line so the
// rendered HTML can restore the original source text exactly.
func normalizeIndentedFences(source []byte) (string, []string) {
	text := string(source)
	lines := strings.Split(text, "\n")
	indents := make([]string, len(lines))
	changed := false

	for i := 0; i < len(lines); i++ {
		body, _ := splitCR(lines[i])
		indent := leadingWhitespace(body)
		if indent == "" {
			continue
		}
		if !openingFence(body[len(indent):]) {
			continue
		}
		end := -1
		for j := i + 1; j < len(lines); j++ {
			candidate, _ := splitCR(lines[j])
			if strings.TrimLeft(candidate, " \t") == "```" {
				end = j
				break
			}
		}
		if end < 0 {
			continue
		}
		for k := i; k <= end; k++ {
			lineBody, lineCR := splitCR(lines[k])
			removed := stripFenceIndent(lineBody, indent)
			lines[k] = lineBody[len(removed):] + lineCR
			indents[k] = removed
		}
		changed = true
		i = end
	}

	if !changed {
		return text, nil
	}
	return strings.Join(lines, "\n"), indents
}

func reindentSourceHTML(highlighted template.HTML, indents []string) template.HTML {
	lines := strings.Split(string(highlighted), "\n")
	for i := range lines {
		if i < len(indents) && indents[i] != "" {
			lines[i] = indents[i] + lines[i]
		}
	}
	return template.HTML(strings.Join(lines, "\n"))
}

func splitCR(line string) (string, string) {
	if strings.HasSuffix(line, "\r") {
		return line[:len(line)-1], "\r"
	}
	return line, ""
}

func leadingWhitespace(line string) string {
	n := 0
	for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
		n++
	}
	return line[:n]
}

func openingFence(line string) bool {
	if !strings.HasPrefix(line, "```") || strings.HasPrefix(line, "````") {
		return false
	}
	info := line[3:]
	return info == "" || (info != "" && !strings.ContainsAny(info, "`"))
}

func stripFenceIndent(line, indent string) string {
	if strings.HasPrefix(line, indent) {
		return indent
	}
	n := 0
	for n < len(line) && n < len(indent) && line[n] == ' ' {
		n++
	}
	return line[:n]
}

func RenderMarkdown(source []byte, sourcePath string, opts RenderOptions) (*Document, error) {
	rawSource := source
	strippedPrefix := 0
	if bom := []byte("\xef\xbb\xbf"); bytes.HasPrefix(source, bom) {
		strippedPrefix = len(bom)
		source = source[len(bom):]
	}
	withoutFrontMatter := stripFrontMatter(source)
	strippedPrefix += len(source) - len(withoutFrontMatter)
	source = withoutFrontMatter
	md := newMarkdown(opts.CurrentDocID)
	docNode := md.Parser().Parse(text.NewReader(source))

	headings := make([]Heading, 0)
	title := ""
	usedIDs := make(map[string]bool)
	err := ast.Walk(docNode, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		textValue := strings.TrimSpace(nodeText(heading, source))
		if textValue == "" {
			textValue = "未命名标题"
		}
		if title == "" && heading.Level == 1 {
			title = textValue
		}

		id := uniqueHeadingID(headingSlug(textValue), usedIDs)
		if opts.IDPrefix != "" {
			id = opts.IDPrefix + "-" + id
		}
		heading.SetAttributeString("id", []byte(id))
		if heading.Level <= 3 {
			line := 0
			if lines := heading.Lines(); lines.Len() > 0 {
				offset := strippedPrefix + lines.At(0).Start
				if offset > len(rawSource) {
					offset = len(rawSource)
				}
				line = bytes.Count(rawSource[:offset], []byte("\n"))
			}
			headings = append(headings, Heading{Level: heading.Level, Text: textValue, ID: id, Line: line})
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	if title == "" {
		title = strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	}

	baseDir := filepath.Dir(sourcePath)
	if abs, err := filepath.Abs(baseDir); err == nil {
		baseDir = abs
	}
	if err := ast.Walk(docNode, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Image:
			rewritten, err := rewriteURL(string(typed.Destination), baseDir, ImageURL, opts.KnownDocs, opts.CurrentDocID, opts.PageLinks)
			if err != nil {
				return ast.WalkStop, err
			}
			typed.Destination = []byte(rewritten)
		case *ast.Link:
			rewritten, err := rewriteURL(string(typed.Destination), baseDir, LinkURL, opts.KnownDocs, opts.CurrentDocID, opts.PageLinks)
			if err != nil {
				return ast.WalkStop, err
			}
			typed.Destination = []byte(rewritten)
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, source, docNode); err != nil {
		return nil, err
	}

	return &Document{
		Path:  sourcePath,
		Title: title,
		HTML:  template.HTML(buf.String()),
		TOC:   headings,
	}, nil
}

func headingSlug(text string) string {
	var b strings.Builder
	pendingDash := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			pendingDash = false
			continue
		}
		if unicode.IsSpace(r) || r == '-' || r == '_' {
			pendingDash = b.Len() > 0
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "section"
	}
	return slug
}

func uniqueHeadingID(base string, used map[string]bool) string {
	if !used[base] {
		used[base] = true
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

func nodeText(node ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := child.(type) {
		case *ast.Text:
			b.Write(typed.Segment.Value(source))
			if typed.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(typed.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}
