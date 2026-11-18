package render

import (
	"fmt"
	stdhtml "html"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	maxLinkedDocs      = 20
	maxLinkedFileBytes = 2 << 20
	maxLinkedTotal     = 10 << 20
)

type sourceDoc struct {
	Path    string
	RelPath string
	Source  []byte
	ReadErr error
	Linked  bool
}

type linkedCandidate struct {
	path    string
	display string
}

func buildPageFromSources(mode, title string, initial []sourceDoc) (*Page, error) {
	linked := collectLinkedMDSources(initial)
	all := make([]sourceDoc, 0, len(initial)+len(linked))
	all = append(all, initial...)
	all = append(all, linked...)

	knownDocs := make(map[string]string, len(all))
	for i, doc := range all {
		knownDocs[pathKey(doc.Path)] = fmt.Sprintf("doc-%d", i)
	}

	docs := make([]Document, 0, len(all))
	for i, sourceDoc := range all {
		id := fmt.Sprintf("doc-%d", i)
		if sourceDoc.ReadErr != nil {
			docs = append(docs, errorDocumentForSource(sourceDoc, id))
			continue
		}
		doc, err := RenderMarkdown(sourceDoc.Source, sourceDoc.Path, RenderOptions{
			IDPrefix:     id,
			CurrentDocID: id,
			KnownDocs:    knownDocs,
		})
		if err != nil {
			docs = append(docs, errorDocumentForSource(sourceDoc, id))
			continue
		}
		doc.ID = id
		doc.Dir = filepath.Dir(sourceDoc.Path)
		doc.RelPath = sourceDoc.RelPath
		doc.Linked = sourceDoc.Linked
		docs = append(docs, *doc)
	}

	return &Page{
		Mode:      mode,
		Title:     title,
		Active:    0,
		Documents: docs,
	}, nil
}

func collectLinkedMDSources(initial []sourceDoc) []sourceDoc {
	known := make(map[string]bool, len(initial))
	for _, doc := range initial {
		known[pathKey(doc.Path)] = true
	}

	md := newMarkdown("")
	seen := make(map[string]bool)
	candidates := make([]linkedCandidate, 0)
	for _, doc := range initial {
		if doc.ReadErr != nil {
			continue
		}
		baseDir := filepath.Dir(doc.Path)
		if abs, err := filepath.Abs(baseDir); err == nil {
			baseDir = abs
		}
		root := md.Parser().Parse(text.NewReader(doc.Source))
		_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			link, ok := node.(*ast.Link)
			if !ok {
				return ast.WalkContinue, nil
			}
			abs, display, ok := resolveLinkedMarkdown(string(link.Destination), baseDir)
			if !ok {
				return ast.WalkContinue, nil
			}
			key := pathKey(abs)
			if known[key] || seen[key] {
				return ast.WalkContinue, nil
			}
			// Never stat or read a network path during initial load.
			if isNetworkPath(abs) {
				return ast.WalkContinue, nil
			}
			seen[key] = true
			candidates = append(candidates, linkedCandidate{path: abs, display: display})
			if len(candidates) >= maxLinkedDocs {
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if len(candidates) >= maxLinkedDocs {
			break
		}
	}

	linked := make([]sourceDoc, 0, len(candidates))
	var total int64
	for _, candidate := range candidates {
		info, err := os.Stat(candidate.path)
		if err != nil || info.IsDir() || info.Size() > maxLinkedFileBytes {
			continue
		}
		if total+info.Size() > maxLinkedTotal {
			break
		}
		source, err := readFileLimited(candidate.path, maxLinkedFileBytes)
		if err != nil {
			continue
		}
		total += info.Size()
		linked = append(linked, sourceDoc{
			Path:    candidate.path,
			RelPath: candidate.display,
			Source:  source,
			Linked:  true,
		})
	}
	return linked
}

func resolveLinkedMarkdown(raw, baseDir string) (absPath, display string, ok bool) {
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", "", false
	}

	target, ok := resolveLocalDestination(raw, baseDir)
	if !ok {
		return "", "", false
	}
	if !isMarkdownPath(target.Path) || isNetworkPath(target.Path) {
		return "", "", false
	}
	display = target.Display
	if display == "" {
		display = target.Path
	}
	return target.Path, display, true
}

func isMarkdownPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func pathKey(path string) string {
	return strings.ToLower(filepath.Clean(path))
}

func errorDocumentForSource(doc sourceDoc, id string) Document {
	name := doc.RelPath
	if name == "" {
		name = filepath.Base(doc.Path)
	}
	title := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	body := fmt.Sprintf(`<p class="read-error">无法读取 <code>%s</code></p>`, stdhtml.EscapeString(name))
	return Document{
		Path:    doc.Path,
		Dir:     filepath.Dir(doc.Path),
		RelPath: name,
		ID:      id,
		Title:   title,
		HTML:    template.HTML(body),
		Linked:  doc.Linked,
	}
}
