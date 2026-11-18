package render

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"openmd/internal/buildinfo"

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

//go:embed web/favicon.png
var faviconPNG []byte

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateSource))

type pageData struct {
	Title            string
	Mode             string
	Active           int
	ActiveID         string
	ActiveTitle      string
	ActivePath       string
	ActiveDir        string
	Documents        []Document
	LinkedDocuments  []Document
	Warnings         []string
	VersionLabel     string
	BuildTime        string
	AuthorName       string
	AuthorID         string
	ReleaseAvailable bool
	ReleaseURL       string
	Favicon          template.URL
	CSS              template.CSS
	MarkdownCSS      template.CSS
	ChromaCSS        template.CSS
	JS               template.JS
}

var maxSingleFileBytes int64 = 32 << 20

func makeAbsolute(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func BuildFilePage(filePath string) (*Page, error) {
	filePath = makeAbsolute(filePath)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %s 是目录", filepath.Base(filePath))
	}
	if maxSingleFileBytes > 0 && info.Size() > maxSingleFileBytes {
		return nil, fmt.Errorf(
			"Markdown 文件过大：%s（%d 字节），上限 %d 字节",
			filepath.Base(filePath), info.Size(), maxSingleFileBytes,
		)
	}
	source, err := readFileLimited(filePath, maxSingleFileBytes)
	if err != nil {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %w", err)
	}
	page, err := buildPageFromSources("file", "", []sourceDoc{{
		Path:    filePath,
		RelPath: filepath.Base(filePath),
		Source:  source,
	}})
	if err != nil {
		return nil, err
	}
	if len(page.Documents) > 0 {
		page.Title = page.Documents[0].Title
	}
	return page, nil
}

func RenderPage(page *Page) ([]byte, error) {
	light, dark := HighlightCSS()
	activeTitle := page.Title
	activeID := ""
	activePath := ""
	activeDir := ""
	if page.Active >= 0 && page.Active < len(page.Documents) {
		activeTitle = page.Documents[page.Active].Title
		activeID = page.Documents[page.Active].ID
		activePath = page.Documents[page.Active].Path
		activeDir = page.Documents[page.Active].Dir
	}
	var linkedDocs []Document
	for _, doc := range page.Documents {
		if doc.Linked {
			linkedDocs = append(linkedDocs, doc)
		}
	}
	data := pageData{
		Title:            page.Title,
		Mode:             page.Mode,
		Active:           page.Active,
		ActiveID:         activeID,
		ActiveTitle:      activeTitle,
		ActivePath:       activePath,
		ActiveDir:        activeDir,
		Documents:        page.Documents,
		LinkedDocuments:  linkedDocs,
		Warnings:         page.Warnings,
		VersionLabel:     buildinfo.VersionLabel(),
		BuildTime:        buildinfo.BuildTimeDisplay(),
		AuthorName:       buildinfo.AuthorName,
		AuthorID:         buildinfo.AuthorID,
		ReleaseAvailable: buildinfo.IsReleaseBuild(),
		ReleaseURL:       buildinfo.ReleaseURL,
		Favicon:          template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(faviconPNG)),
		CSS:              template.CSS(readerCSS),
		MarkdownCSS:      template.CSS(githubMarkdownCSS),
		ChromaCSS:        template.CSS(light + "\n" + dark),
		JS:               template.JS(readerJS),
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
