package render

import (
	"fmt"
	stdhtml "html"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"openmd/internal/scan"
)

func BuildDirPage(dirPath string) (*Page, error) {
	result, err := scan.Directory(dirPath, scan.DefaultLimits)
	if err != nil {
		return nil, err
	}
	page, err := buildDirPage(result.Documents)
	if err != nil {
		return nil, err
	}
	page.Title = filepath.Base(filepath.Clean(dirPath))
	page.Warnings = append(page.Warnings, result.Warnings...)
	if result.Truncated {
		page.Warnings = append(page.Warnings, fmt.Sprintf("目录内容过多，只加载了前 %d 篇", len(result.Documents)))
	}
	return page, nil
}

func buildDirPage(docs []scan.Document) (*Page, error) {
	knownDocs := make(map[string]string, len(docs))
	for i, doc := range docs {
		knownDocs[strings.ToLower(filepath.Clean(doc.Path))] = fmt.Sprintf("doc-%d", i)
	}

	rendered := make([]Document, 0, len(docs))
	for i, doc := range docs {
		id := fmt.Sprintf("doc-%d", i)
		source, err := os.ReadFile(doc.Path)
		if err != nil {
			rendered = append(rendered, errorDocument(doc, id))
			continue
		}
		renderedDoc, err := RenderMarkdown(source, doc.Path, RenderOptions{
			IDPrefix:  id,
			KnownDocs: knownDocs,
		})
		if err != nil {
			rendered = append(rendered, errorDocument(doc, id))
			continue
		}
		renderedDoc.ID = id
		renderedDoc.RelPath = doc.Name
		rendered = append(rendered, *renderedDoc)
	}

	return &Page{
		Mode:      "dir",
		Active:    0,
		Documents: rendered,
	}, nil
}

func errorDocument(doc scan.Document, id string) Document {
	title := strings.TrimSuffix(doc.Name, filepath.Ext(doc.Name))
	body := fmt.Sprintf(`<p class="read-error">无法读取 <code>%s</code></p>`, stdhtml.EscapeString(doc.Name))
	return Document{
		Path:    doc.Path,
		RelPath: doc.Name,
		ID:      id,
		Title:   title,
		HTML:    template.HTML(body),
	}
}
