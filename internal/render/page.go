package render

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

type Page struct {
	Mode      string
	Title     string
	Active    int
	Documents []Document
	Warnings  []string
}

//go:embed web/page.html
var pageTemplateSource string

//go:embed web/reader.css
var readerCSS string

//go:embed web/reader.js
var readerJS string

//go:embed web/vendor/github-markdown/github-markdown.css
var githubMarkdownCSS string

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateSource))

type pageData struct {
	Title       string
	Mode        string
	Active      int
	ActiveTitle string
	Documents   []Document
	Warnings    []string
	CSS         template.CSS
	MarkdownCSS template.CSS
	ChromaCSS   template.CSS
	JS          template.JS
}

func BuildFilePage(filePath string) (*Page, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %w", err)
	}
	doc, err := RenderMarkdown(source, filePath, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		return nil, err
	}
	doc.ID = "doc-0"
	doc.RelPath = filepath.Base(filePath)
	return &Page{
		Mode:      "file",
		Title:     doc.Title,
		Active:    0,
		Documents: []Document{*doc},
	}, nil
}

func RenderPage(page *Page) ([]byte, error) {
	light, dark := HighlightCSS()
	activeTitle := page.Title
	if page.Active >= 0 && page.Active < len(page.Documents) {
		activeTitle = page.Documents[page.Active].Title
	}
	data := pageData{
		Title:       page.Title,
		Mode:        page.Mode,
		Active:      page.Active,
		ActiveTitle: activeTitle,
		Documents:   page.Documents,
		Warnings:    page.Warnings,
		CSS:         template.CSS(readerCSS),
		MarkdownCSS: template.CSS(githubMarkdownCSS),
		ChromaCSS:   template.CSS(light + "\n" + dark),
		JS:          template.JS(readerJS),
	}
	var buf bytes.Buffer
	if err := pageTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HighlightCSS() (light, dark string) {
	return chromaCSS("github"), chromaCSS("github-dark")
}

func chromaCSS(styleName string) string {
	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.WithModeClasses(true),
	)
	var buf strings.Builder
	if err := formatter.WriteCSS(&buf, styles.Get(styleName)); err != nil {
		return ""
	}
	return buf.String()
}
