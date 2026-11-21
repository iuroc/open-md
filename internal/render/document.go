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
	lexer := lexers.Get("markdown")
	if lexer != nil {
		lexer = chroma.Coalesce(lexer)
	}
	return highlightSourceWith(lexer, source)
}

// sourceHighlighter is replaceable in tests to exercise the fallback path.
var sourceHighlighter = highlightSource

func highlightSourceWith(lexer chroma.Lexer, source []byte) template.HTML {
	if len(source) == 0 {
		return template.HTML("")
	}
	fallback := func() template.HTML {
		return template.HTML(stdhtml.EscapeString(string(source)))
	}
	if lexer == nil {
		return fallback()
	}
	iterator, err := lexer.Tokenise(nil, string(source))
	if err != nil || iterator == nil {
		return fallback()
	}
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.WithModeClasses(true),
		chromahtml.PreventSurroundingPre(true),
	)
	var buf bytes.Buffer
	if err := formatter.Format(&buf, styles.Get("github"), iterator); err != nil {
		return fallback()
	}
	return template.HTML(buf.String())
}

func RenderMarkdown(source []byte, sourcePath string, opts RenderOptions) (*Document, error) {
	source = bytes.TrimPrefix(source, []byte("\xef\xbb\xbf"))
	source = stripFrontMatter(source)
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
			headings = append(headings, Heading{Level: heading.Level, Text: textValue, ID: id})
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
			rewritten, err := RewriteURL(string(typed.Destination), baseDir, ImageURL, opts.KnownDocs, opts.CurrentDocID)
			if err != nil {
				return ast.WalkStop, err
			}
			typed.Destination = []byte(rewritten)
		case *ast.Link:
			rewritten, err := RewriteURL(string(typed.Destination), baseDir, LinkURL, opts.KnownDocs, opts.CurrentDocID)
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
