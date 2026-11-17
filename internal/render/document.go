package render

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
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
	Path    string
	RelPath string
	ID      string
	Title   string
	HTML    template.HTML
	TOC     []Heading
}

type RenderOptions struct {
	IDPrefix  string
	KnownDocs map[string]string
}

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithModeClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(util.Prioritized(&safeLinkRenderer{}, 100)),
		),
	)
}

func RenderMarkdown(source []byte, sourcePath string, opts RenderOptions) (*Document, error) {
	source = bytes.TrimPrefix(source, []byte("\xef\xbb\xbf"))
	md := newMarkdown()
	docNode := md.Parser().Parse(text.NewReader(source))

	headings := make([]Heading, 0)
	title := ""
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

		id := headingID(heading)
		if id == "" {
			id = "heading-" + strings.TrimSpace(strings.ToLower(strings.ReplaceAll(textValue, " ", "-")))
			if id == "heading-" {
				id = "heading-" + "section"
			}
		}
		if opts.IDPrefix != "" {
			id = opts.IDPrefix + "-" + id
			heading.SetAttributeString("id", []byte(id))
		} else if headingID(heading) == "" {
			heading.SetAttributeString("id", []byte(id))
		}
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
			rewritten, err := RewriteURL(string(typed.Destination), baseDir, ImageURL, opts.KnownDocs)
			if err != nil {
				return ast.WalkStop, err
			}
			typed.Destination = []byte(rewritten)
		case *ast.Link:
			rewritten, err := RewriteURL(string(typed.Destination), baseDir, LinkURL, opts.KnownDocs)
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

func headingID(heading *ast.Heading) string {
	value, ok := heading.AttributeString("id")
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case string:
		return typed
	default:
		return ""
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
