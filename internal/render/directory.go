package render

import (
	"fmt"
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
	page.Active = initialDocIndex(result.Documents)
	page.Tree = buildTree(page.Documents[:len(result.Documents)], page.Documents[page.Active].ID)
	return page, nil
}

func buildDirPage(docs []scan.Document) (*Page, error) {
	initial := make([]sourceDoc, 0, len(docs))
	for _, doc := range docs {
		source, err := readFileLimited(doc.Path, scan.DefaultLimits.MaxTotalBytes)
		rel := doc.RelPath
		if rel == "" {
			rel = doc.Name
		}
		initial = append(initial, sourceDoc{
			Path:    doc.Path,
			RelPath: rel,
			Source:  source,
			ReadErr: err,
		})
	}
	return buildPageFromSources("dir", "", initial)
}

func initialDocIndex(docs []scan.Document) int {
	for _, target := range []string{"readme.md", "index.md", "readme.markdown", "index.markdown"} {
		for i, doc := range docs {
			if strings.ToLower(doc.RelPath) == target {
				return i
			}
		}
	}
	return 0
}
