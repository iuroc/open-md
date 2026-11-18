package render

import (
	"fmt"
	"os"
	"path/filepath"

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
	initial := make([]sourceDoc, 0, len(docs))
	for _, doc := range docs {
		source, err := os.ReadFile(doc.Path)
		initial = append(initial, sourceDoc{
			Path:    doc.Path,
			RelPath: doc.Name,
			Source:  source,
			ReadErr: err,
		})
	}
	return buildPageFromSources("dir", "", initial)
}
