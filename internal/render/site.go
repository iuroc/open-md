package render

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"openmd/internal/scan"
)

const (
	siteFullThresholdDocs  = 3000
	siteFullThresholdBytes = 8 << 20
)

var ErrLoadCancelled = errors.New("用户取消加载")

type LoadChoice int

const (
	LoadAll LoadChoice = iota
	LoadQuick
	LoadCancel
)

type Estimate struct {
	TotalDocs    int
	TotalBytes   int64
	PendingDocs  int
	PendingBytes int64
}

type SiteOptions struct {
	CacheBase          string
	Version            string
	AllowCache         bool
	Confirm            func(Estimate) LoadChoice
	FullThresholdDocs  int
	FullThresholdBytes int64
}

func (o SiteOptions) thresholdDocs() int {
	if o.FullThresholdDocs > 0 {
		return o.FullThresholdDocs
	}
	return siteFullThresholdDocs
}

func (o SiteOptions) thresholdBytes() int64 {
	if o.FullThresholdBytes > 0 {
		return o.FullThresholdBytes
	}
	return siteFullThresholdBytes
}

type Site struct {
	InitialPage string
	Generated   int
	Reused      int
	Warnings    []string
	Pages       int
}

type siteLink struct {
	Path    string `json:"path"`
	Display string `json:"display"`
}

type siteDoc struct {
	Path       string
	RelPath    string
	Href       string
	Title      string
	Source     []byte
	ReadErr    error
	MTime      int64
	Size       int64
	Linked     bool
	Error      bool
	Links      []siteLink
	NeedRender bool
	Inline     bool
}

type siteManifest struct {
	Version   string            `json:"version"`
	KnownHash string            `json:"knownHash"`
	Root      string            `json:"root"`
	Docs      []siteManifestDoc `json:"docs"`
}

type siteManifestDoc struct {
	Path    string     `json:"path"`
	Href    string     `json:"href"`
	Title   string     `json:"title"`
	MTime   int64      `json:"mtime"`
	Size    int64      `json:"size"`
	Linked  bool       `json:"linked,omitempty"`
	Links   []siteLink `json:"links,omitempty"`
	Errored bool       `json:"errored,omitempty"`
	Inline  bool       `json:"inline,omitempty"`
}

func BuildDirSite(dirPath string, opts SiteOptions) (*Site, error) {
	abs, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("无法解析目录路径: %w", err)
	}
	abs = filepath.Clean(abs)
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("无法读取目录: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("无法读取目录: %s 不是目录", filepath.Base(abs))
	}
	if opts.Version == "" {
		opts.Version = "dev"
	}
	rootKey := siteRootKey(abs)
	cacheDir := filepath.Join(opts.CacheBase, rootKey)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	manifest := loadManifest(cacheDir)

	result, err := scan.Directory(abs, scan.UnlimitedLimits)
	if err != nil {
		return nil, err
	}
	warnings := append([]string{}, result.Warnings...)

	pendingDocs, pendingBytes := pendingWork(result.Documents, manifest, cacheDir)
	if !opts.AllowCache || manifest.Version != opts.Version || manifest.Root != abs {
		pendingDocs = len(result.Documents)
		pendingBytes = totalBytes(result.Documents)
	}
	if pendingDocs > opts.thresholdDocs() || pendingBytes > opts.thresholdBytes() {
		choice := LoadAll
		if opts.Confirm != nil {
			choice = opts.Confirm(Estimate{
				TotalDocs:    len(result.Documents),
				TotalBytes:   totalBytes(result.Documents),
				PendingDocs:  pendingDocs,
				PendingBytes: pendingBytes,
			})
		}
		switch choice {
		case LoadCancel:
			return nil, ErrLoadCancelled
		case LoadQuick:
			quick, quickErr := scan.Directory(abs, scan.QuickLimits)
			if quickErr != nil {
				return nil, quickErr
			}
			result = quick
			warnings = append(warnings, quick.Warnings...)
			if quick.Truncated {
				warnings = append(warnings, fmt.Sprintf("已加载前 %d 篇；其余文档未加载", len(quick.Documents)))
			}
		}
	}

	primary := make([]*siteDoc, 0, len(result.Documents))
	entryByPath := manifestByPath(manifest)
	primaryPaths := make(map[string]bool, len(result.Documents))
	for _, doc := range result.Documents {
		primaryPaths[pathKey(doc.Path)] = true
		primary = append(primary, &siteDoc{
			Path:    doc.Path,
			RelPath: doc.RelPath,
			MTime:   doc.ModTime,
			Size:    doc.Size,
		})
	}
	active := initialSiteDocIndex(primary)
	primary[active].Inline = true
	for _, sd := range primary {
		entry, ok := entryByPath[pathKey(sd.Path)]
		if ok && entry.MTime == sd.MTime && entry.Size == sd.Size && cacheDocReady(cacheDir, entry.Href) && entry.Inline == sd.Inline {
			sd.Title = entry.Title
			sd.Links = entry.Links
			sd.Href = entry.Href
		} else {
			sd.NeedRender = true
			readPrimary(sd)
		}
	}

	linked, linkedWarnings := collectLinkedDocs(primary, primaryPaths, entryByPath, cacheDir, abs)
	warnings = append(warnings, linkedWarnings...)
	all := append(append([]*siteDoc{}, primary...), linked...)
	assignHrefs(all)
	knownDocs := make(map[string]string, len(all))
	for _, doc := range all {
		knownDocs[pathKey(doc.Path)] = doc.Href
	}
	knownHash := knownSetHash(all)
	rebuildAll := !opts.AllowCache || manifest.Version != opts.Version || manifest.KnownHash != knownHash || manifest.Root != abs
	if rebuildAll {
		for _, doc := range all {
			doc.NeedRender = true
		}
	}

	hasFilesPanel := true
	site := &Site{Warnings: warnings, Pages: len(all)}
	if err := renderDocs(cacheDir, "dir", rootKey, all, knownDocs, hasFilesPanel); err != nil {
		return nil, err
	}

	treeDocs := make([]Document, 0, len(primary))
	for _, doc := range primary {
		treeDocs = append(treeDocs, Document{ID: doc.Href, Title: doc.Title, RelPath: doc.RelPath, Linked: false})
	}
	linkedDocs := make([]Document, 0, len(linked))
	for _, doc := range linked {
		linkedDocs = append(linkedDocs, Document{ID: doc.Href, Title: doc.Title, RelPath: doc.RelPath, Linked: true})
	}
	navHTML := renderNavHTML(buildTree(treeDocs, ""), linkedDocs, warnings, "dir")
	if err := writeTreeJS(cacheDir, navHTML); err != nil {
		return nil, err
	}
	if err := writeAssets(cacheDir); err != nil {
		return nil, err
	}
	if err := writeManifest(cacheDir, manifestFor(opts.Version, knownHash, abs, all)); err != nil {
		return nil, err
	}
	touchDir(cacheDir)
	prunePages(cacheDir, all)

	site.InitialPage = filepath.Join(cacheDir, primary[active].Href)
	for _, doc := range all {
		if doc.NeedRender {
			site.Generated++
		} else {
			site.Reused++
		}
	}
	return site, nil
}

func BuildFileSite(filePath string, opts SiteOptions) (*Site, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法解析文件路径: %w", err)
	}
	abs = filepath.Clean(abs)
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("无法读取 Markdown 文件: %s 是目录", filepath.Base(abs))
	}
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if info.Size() > 32<<20 {
		choice := LoadAll
		if opts.Confirm != nil {
			choice = opts.Confirm(Estimate{TotalDocs: 1, TotalBytes: info.Size(), PendingDocs: 1, PendingBytes: info.Size()})
		}
		if choice == LoadCancel {
			return nil, ErrLoadCancelled
		}
	}

	rootKey := siteRootKey(abs)
	cacheDir := filepath.Join(opts.CacheBase, rootKey)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	manifest := loadManifest(cacheDir)
	entryByPath := manifestByPath(manifest)

	primaryDoc := &siteDoc{
		Path:    abs,
		RelPath: filepath.Base(abs),
		MTime:   info.ModTime().UnixNano(),
		Size:    info.Size(),
		Inline:  true,
	}
	if entry, ok := entryByPath[pathKey(abs)]; ok && entry.MTime == primaryDoc.MTime && entry.Size == primaryDoc.Size && cacheDocReady(cacheDir, entry.Href) && entry.Inline == primaryDoc.Inline {
		primaryDoc.Title = entry.Title
		primaryDoc.Links = entry.Links
		primaryDoc.Href = entry.Href
	} else {
		primaryDoc.NeedRender = true
		readPrimary(primaryDoc)
	}
	primary := []*siteDoc{primaryDoc}
	primaryPaths := map[string]bool{pathKey(abs): true}
	linked, warnings := collectLinkedDocs(primary, primaryPaths, entryByPath, cacheDir, filepath.Dir(abs))
	all := append(append([]*siteDoc{}, primary...), linked...)
	assignHrefs(all)
	knownDocs := make(map[string]string, len(all))
	for _, doc := range all {
		knownDocs[pathKey(doc.Path)] = doc.Href
	}
	knownHash := knownSetHash(all)
	rebuildAll := !opts.AllowCache || manifest.Version != opts.Version || manifest.KnownHash != knownHash || manifest.Root != abs
	if rebuildAll {
		for _, doc := range all {
			doc.NeedRender = true
		}
	}
	hasFilesPanel := len(linked) > 0
	if err := renderDocs(cacheDir, "file", rootKey, all, knownDocs, hasFilesPanel); err != nil {
		return nil, err
	}
	linkedDocs := make([]Document, 0, len(linked))
	for _, doc := range linked {
		linkedDocs = append(linkedDocs, Document{ID: doc.Href, Title: doc.Title, RelPath: doc.RelPath, Linked: true})
	}
	navHTML := renderNavHTML(nil, linkedDocs, warnings, "file")
	if err := writeTreeJS(cacheDir, navHTML); err != nil {
		return nil, err
	}
	if err := writeAssets(cacheDir); err != nil {
		return nil, err
	}
	if err := writeManifest(cacheDir, manifestFor(opts.Version, knownHash, abs, all)); err != nil {
		return nil, err
	}
	touchDir(cacheDir)
	prunePages(cacheDir, all)

	site := &Site{InitialPage: filepath.Join(cacheDir, primaryDoc.Href), Warnings: warnings, Pages: len(all)}
	for _, doc := range all {
		if doc.NeedRender {
			site.Generated++
		} else {
			site.Reused++
		}
	}
	return site, nil
}

func readPrimary(doc *siteDoc) {
	source, err := os.ReadFile(doc.Path)
	doc.Source = source
	doc.ReadErr = err
	if err == nil {
		doc.Links = extractMarkdownLinks(source, filepath.Dir(doc.Path))
	}
	if doc.Title == "" {
		doc.Title = strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	}
}

func collectLinkedDocs(primary []*siteDoc, primaryPaths map[string]bool, entries map[string]siteManifestDoc, cacheDir, displayBase string) ([]*siteDoc, []string) {
	seen := make(map[string]bool)
	targets := make([]siteLink, 0)
	var total int64
	warnings := make([]string, 0)
	for _, doc := range primary {
		for _, link := range doc.Links {
			key := pathKey(link.Path)
			if primaryPaths[key] || seen[key] || isNetworkPath(link.Path) {
				continue
			}
			if len(targets) >= maxLinkedDocs || total > maxLinkedTotal {
				break
			}
			info, err := os.Stat(link.Path)
			if err != nil || info.IsDir() || info.Size() > maxLinkedFileBytes {
				continue
			}
			if total+info.Size() > maxLinkedTotal {
				continue
			}
			total += info.Size()
			seen[key] = true
			targets = append(targets, link)
		}
	}
	linked := make([]*siteDoc, 0, len(targets))
	for _, target := range targets {
		info, err := os.Stat(target.Path)
		if err != nil || info.IsDir() {
			warnings = append(warnings, "链接文档无法读取："+target.Display)
			continue
		}
		rel, relErr := filepath.Rel(displayBase, target.Path)
		if relErr != nil || rel == "" {
			rel = target.Display
		}
		sd := &siteDoc{
			Path:    target.Path,
			RelPath: filepath.ToSlash(rel),
			MTime:   info.ModTime().UnixNano(),
			Size:    info.Size(),
			Linked:  true,
		}
		if sd.RelPath == "" {
			sd.RelPath = filepath.Base(target.Path)
		}
		entry, ok := entries[pathKey(target.Path)]
		if ok && entry.MTime == sd.MTime && entry.Size == sd.Size && cacheDocReady(cacheDir, entry.Href) && entry.Inline == sd.Inline {
			sd.Title = entry.Title
			sd.Href = entry.Href
		} else {
			sd.NeedRender = true
			source, readErr := os.ReadFile(sd.Path)
			sd.Source = source
			sd.ReadErr = readErr
			sd.Title = strings.TrimSuffix(filepath.Base(sd.Path), filepath.Ext(sd.Path))
		}
		linked = append(linked, sd)
	}
	return linked, warnings
}

func extractMarkdownLinks(source []byte, baseDir string) []siteLink {
	md := newMarkdown("")
	root := md.Parser().Parse(text.NewReader(source))
	seen := make(map[string]bool)
	links := make([]siteLink, 0)
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := node.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		abs, display, ok := resolveLinkedMarkdown(string(link.Destination), baseDir)
		if !ok || isNetworkPath(abs) {
			return ast.WalkContinue, nil
		}
		key := pathKey(abs)
		if seen[key] {
			return ast.WalkContinue, nil
		}
		seen[key] = true
		links = append(links, siteLink{Path: abs, Display: display})
		return ast.WalkContinue, nil
	})
	return links
}

func renderDocs(cacheDir, mode, rootKey string, docs []*siteDoc, knownDocs map[string]string, hasFilesPanel bool) error {
	jobs := make(chan *siteDoc)
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > 8 {
		workers = 8
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for doc := range jobs {
				if !doc.NeedRender {
					continue
				}
				if err := renderOneDoc(cacheDir, mode, rootKey, doc, knownDocs, hasFilesPanel); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}
		}()
	}
	for _, doc := range docs {
		jobs <- doc
	}
	close(jobs)
	wg.Wait()
	return firstErr
}

func renderOneDoc(cacheDir, mode, rootKey string, doc *siteDoc, knownDocs map[string]string, hasFilesPanel bool) error {
	if doc.Source == nil && doc.ReadErr == nil {
		doc.Source, doc.ReadErr = os.ReadFile(doc.Path)
	}
	document := Document{
		Path:    doc.Path,
		Dir:     filepath.Dir(doc.Path),
		RelPath: doc.RelPath,
		ID:      "doc",
		Linked:  doc.Linked,
	}
	if doc.ReadErr != nil {
		document.Title = doc.Title
		if document.Title == "" {
			document.Title = strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
		}
		document.HTML = template.HTML(fmt.Sprintf(`<p class="read-error">无法读取 <code>%s</code></p>`, stdhtml.EscapeString(doc.RelPath)))
		doc.Title = document.Title
		doc.Error = true
	} else {
		rendered, err := RenderMarkdown(doc.Source, doc.Path, RenderOptions{
			KnownDocs: knownDocs,
			PageLinks: true,
		})
		if err != nil {
			document.Title = strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
			document.HTML = template.HTML(`<p class="read-error">无法渲染该文档</p>`)
			doc.Title = document.Title
			doc.Error = true
		} else {
			rendered.ID = "doc"
			rendered.RelPath = doc.RelPath
			rendered.Linked = doc.Linked
			rendered.HasSource = true
			rendered.SourceHTML = highlightSource(doc.Source)
			document = *rendered
			document.Dir = filepath.Dir(doc.Path)
			doc.Title = rendered.Title
		}
	}
	data := DocPageData{
		Title:         document.Title,
		Mode:          mode,
		RootKey:       rootKey,
		Href:          doc.Href,
		Payload:       docPayloadHref(doc.Href),
		InlineSection: doc.Inline,
		HasFilesPanel: hasFilesPanel,
		Doc:           document,
	}
	html, err := RenderDocPage(data)
	if err != nil {
		return err
	}
	section, err := RenderDocSection(data)
	if err != nil {
		return err
	}
	doc.Title = document.Title
	if err := writeAtomic(filepath.Join(cacheDir, doc.Href), html); err != nil {
		return err
	}
	return writeDocPayload(cacheDir, doc.Href, document, section)
}

func pendingWork(docs []scan.Document, manifest siteManifest, cacheDir string) (int, int64) {
	entries := manifestByPath(manifest)
	pending := 0
	var bytes int64
	for _, doc := range docs {
		entry, ok := entries[pathKey(doc.Path)]
		if !ok || entry.MTime != doc.ModTime || entry.Size != doc.Size || !cacheDocReady(cacheDir, entry.Href) {
			pending++
			bytes += doc.Size
		}
	}
	for _, entry := range manifest.Docs {
		if !entry.Linked {
			continue
		}
		info, err := os.Stat(entry.Path)
		if err != nil || info.ModTime().UnixNano() != entry.MTime || info.Size() != entry.Size || !cacheDocReady(cacheDir, entry.Href) {
			pending++
			bytes += entry.Size
		}
	}
	return pending, bytes
}

func totalBytes(docs []scan.Document) int64 {
	var total int64
	for _, doc := range docs {
		total += doc.Size
	}
	return total
}

func assignHrefs(docs []*siteDoc) {
	for _, doc := range docs {
		if doc.Href == "" {
			doc.Href = docHref(doc.Path)
		}
	}
}

func knownSetHash(docs []*siteDoc) string {
	keys := make([]string, 0, len(docs))
	for _, doc := range docs {
		keys = append(keys, pathKey(doc.Path)+"|"+doc.Href)
	}
	sort.Strings(keys)
	sum := sha256.Sum256([]byte(strings.Join(keys, "\n")))
	return hex.EncodeToString(sum[:])
}

func manifestFor(version, knownHash, root string, docs []*siteDoc) siteManifest {
	m := siteManifest{Version: version, KnownHash: knownHash, Root: root}
	for _, doc := range docs {
		m.Docs = append(m.Docs, siteManifestDoc{
			Path:    doc.Path,
			Href:    doc.Href,
			Title:   doc.Title,
			MTime:   doc.MTime,
			Size:    doc.Size,
			Linked:  doc.Linked,
			Links:   doc.Links,
			Errored: doc.Error,
			Inline:  doc.Inline,
		})
	}
	return m
}

func manifestByPath(manifest siteManifest) map[string]siteManifestDoc {
	m := make(map[string]siteManifestDoc, len(manifest.Docs))
	for _, doc := range manifest.Docs {
		m[pathKey(doc.Path)] = doc
	}
	return m
}

func loadManifest(cacheDir string) siteManifest {
	data, err := os.ReadFile(filepath.Join(cacheDir, "manifest.json"))
	if err != nil {
		return siteManifest{}
	}
	var manifest siteManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return siteManifest{}
	}
	return manifest
}

func writeManifest(cacheDir string, manifest siteManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(cacheDir, "manifest.json"), data)
}

func writeTreeJS(cacheDir, navHTML string) error {
	data, err := json.Marshal(navHTML)
	if err != nil {
		return err
	}
	content := "window.OPENMD_TREE_HTML = " + string(data) + ";\n"
	return writeAtomic(filepath.Join(cacheDir, "tree.js"), []byte(content))
}

func writeAssets(cacheDir string) error {
	light, dark := HighlightCSS()
	assets := map[string]string{
		"reader.css":          readerCSS,
		"markdown.css":        githubMarkdownCSS,
		"chroma.css":          light + "\n" + dark,
		"reader.js":           readerJS,
	}
	for name, content := range assets {
		if err := writeAtomic(filepath.Join(cacheDir, name), []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func touchDir(path string) {
	now := time.Now()
	_ = os.Chtimes(path, now, now)
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func docPayloadHref(href string) string {
	return strings.TrimSuffix(href, ".html") + ".js"
}

func cacheDocReady(cacheDir, href string) bool {
	return fileExists(filepath.Join(cacheDir, href)) && fileExists(filepath.Join(cacheDir, docPayloadHref(href)))
}

type docPayload struct {
	Title       string `json:"title"`
	Path        string `json:"path"`
	Dir         string `json:"dir"`
	RelPath     string `json:"relPath"`
	SectionHTML string `json:"sectionHTML"`
}

func writeDocPayload(cacheDir, href string, doc Document, section []byte) error {
	key, err := json.Marshal(docPayloadHref(href))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("window.OPENMD_PAGES=window.OPENMD_PAGES||{};window.OPENMD_PAGES[")
	buf.Write(key)
	buf.WriteString("]=")
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(docPayload{
		Title:       doc.Title,
		Path:        doc.Path,
		Dir:         doc.Dir,
		RelPath:     doc.RelPath,
		SectionHTML: string(section),
	}); err != nil {
		return err
	}
	return writeAtomic(filepath.Join(cacheDir, docPayloadHref(href)), buf.Bytes())
}

func prunePages(cacheDir string, docs []*siteDoc) {
	keep := make(map[string]bool, len(docs))
	for _, doc := range docs {
		keep[doc.Href] = true
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "doc-") {
			continue
		}
		var htmlName string
		switch {
		case strings.HasSuffix(name, ".html"):
			htmlName = name
		case strings.HasSuffix(name, ".js"):
			htmlName = strings.TrimSuffix(name, ".js") + ".html"
		default:
			continue
		}
		if !keep[htmlName] {
			_ = os.Remove(filepath.Join(cacheDir, name))
		}
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func siteRootKey(path string) string {
	abs := filepath.Clean(path)
	base := sanitizeSiteName(filepath.Base(abs))
	sum := sha256.Sum256([]byte(strings.ToLower(abs)))
	return base + "-" + hex.EncodeToString(sum[:6])
}

func sanitizeSiteName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
		if b.Len() >= 40 {
			break
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "root"
	}
	return result
}

func docHref(path string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(path))))
	return "doc-" + hex.EncodeToString(sum[:8]) + ".html"
}

func initialSiteDocIndex(docs []*siteDoc) int {
	for _, target := range []string{"readme.md", "index.md", "readme.markdown", "index.markdown"} {
		for i, doc := range docs {
			if strings.ToLower(doc.RelPath) == target {
				return i
			}
		}
	}
	return 0
}

func renderNavHTML(tree []*TreeNode, linked []Document, warnings []string, mode string) string {
	var b strings.Builder
	if len(tree) > 0 {
		b.WriteString(`<ul class="tree-root">`)
		renderTreeNodes(&b, tree)
		b.WriteString(`</ul>`)
	}
	if len(linked) > 0 {
		b.WriteString(`<div class="tree-section-label">链接文档</div><ul class="tree-root linked-tree">`)
		for _, doc := range linked {
			b.WriteString(`<li class="tree-node tree-file"><a class="doc-item linked" draggable="false" href="`)
			b.WriteString(stdhtml.EscapeString(doc.ID))
			b.WriteString(`" data-tree-path="`)
			b.WriteString(stdhtml.EscapeString(doc.RelPath))
			b.WriteString(`" title="`)
			b.WriteString(stdhtml.EscapeString(doc.RelPath))
			b.WriteString(`"><span class="doc-title">`)
			b.WriteString(stdhtml.EscapeString(doc.Title))
			b.WriteString(`</span><span class="doc-path">`)
			b.WriteString(stdhtml.EscapeString(doc.RelPath))
			b.WriteString(`</span></a></li>`)
		}
		b.WriteString(`</ul>`)
	}
	if len(warnings) > 0 {
		b.WriteString(`<div class="tree-warning">`)
		for _, warning := range warnings {
			b.WriteString(`<p>`)
			b.WriteString(stdhtml.EscapeString(warning))
			b.WriteString(`</p>`)
		}
		b.WriteString(`</div>`)
	}
	return b.String()
}

func renderTreeNodes(b *strings.Builder, nodes []*TreeNode) {
	for _, node := range nodes {
		if node.IsDir {
			b.WriteString(`<li class="tree-node tree-dir"><button type="button" class="tree-toggle" data-tree-folder data-tree-path="`)
			b.WriteString(stdhtml.EscapeString(node.RelPath))
			b.WriteString(`" aria-expanded="false" title="`)
			b.WriteString(stdhtml.EscapeString(node.RelPath))
			b.WriteString(`"><span class="tree-caret" aria-hidden="true"></span><span class="tree-label">`)
			b.WriteString(stdhtml.EscapeString(node.Name))
			b.WriteString(`</span></button><ul class="tree-children" hidden>`)
			renderTreeNodes(b, node.Children)
			b.WriteString(`</ul></li>`)
			continue
		}
		b.WriteString(`<li class="tree-node tree-file"><a class="tree-file-link" draggable="false" href="`)
		b.WriteString(stdhtml.EscapeString(node.DocID))
		b.WriteString(`" data-tree-path="`)
		b.WriteString(stdhtml.EscapeString(node.RelPath))
		b.WriteString(`" title="`)
		if node.Title != "" {
			b.WriteString(stdhtml.EscapeString(node.Title))
			b.WriteString(` — `)
		}
		b.WriteString(stdhtml.EscapeString(node.RelPath))
		b.WriteString(`"><span class="tree-label">`)
		b.WriteString(stdhtml.EscapeString(node.Name))
		b.WriteString(`</span></a></li>`)
	}
}
