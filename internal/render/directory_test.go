package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openmd/internal/scan"
)

func TestBuildDirPageAndCrossDocumentLink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n\n[B](b.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("# B\n\n## Usage\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildDirPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if page.Mode != "dir" || len(page.Documents) != 2 {
		t.Fatalf("page = %#v", page)
	}
	if !strings.Contains(string(page.Documents[0].HTML), "#doc=doc-1") {
		t.Fatalf("cross-document link not rewritten: %s", page.Documents[0].HTML)
	}
	if !strings.Contains(string(page.Documents[1].HTML), `id="doc-1-usage"`) {
		t.Fatalf("heading ID not prefixed: %s", page.Documents[1].HTML)
	}
}

func TestBuildDirPageKeepsGoingAfterUnreadableDocument(t *testing.T) {
	docs := []scan.Document{
		{Path: filepath.Join(t.TempDir(), "missing.md"), Name: "missing.md"},
		{Path: writeTempMD(t, "# OK"), Name: "ok.md"},
	}
	page, err := buildDirPage(docs)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d", len(page.Documents))
	}
	if !strings.Contains(string(page.Documents[0].HTML), "无法读取") {
		t.Fatal("missing error placeholder")
	}
}

func TestRenderPageShowsScanWarnings(t *testing.T) {
	page := &Page{
		Mode:      "dir",
		Title:     "docs",
		Warnings:  []string{"只加载了前 500 篇"},
		Documents: []Document{{ID: "doc-0", Title: "A"}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "只加载了前 500 篇") {
		t.Fatal("warning not shown")
	}
}

func writeTempMD(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ok.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeRel(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildDirPageRecursiveTreeAndRootReadme(t *testing.T) {
	dir := t.TempDir()
	writeRel(t, dir, "README.md", "# Home\n")
	writeRel(t, dir, "specs/source-view/spec.md", "# source-view\n")
	writeRel(t, dir, "changes/archive/foo/proposal.md", "# proposal\n")

	page, err := BuildDirPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 3 {
		t.Fatalf("docs = %d", len(page.Documents))
	}
	if page.Documents[page.Active].RelPath != "README.md" {
		t.Fatalf("active = %q", page.Documents[page.Active].RelPath)
	}
	if len(page.Tree) == 0 {
		t.Fatal("tree was not built")
	}
	if page.Tree[0].Name != "changes" || page.Tree[1].Name != "specs" || page.Tree[2].Name != "README.md" {
		t.Fatalf("tree = %#v", page.Tree)
	}
}

func TestBuildDirPageInitialDocumentFallsBackToShallowest(t *testing.T) {
	dir := t.TempDir()
	writeRel(t, dir, "specs/source-view/spec.md", "# source-view\n")
	writeRel(t, dir, "changes/archive/foo/proposal.md", "# proposal\n")

	page, err := BuildDirPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if page.Documents[page.Active].RelPath != "specs/source-view/spec.md" {
		t.Fatalf("active = %q", page.Documents[page.Active].RelPath)
	}
}

func TestBuildDirPageKeepsLinkedOutsideDocumentsOutOfTree(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "docs")
	writeRel(t, base, "outside.md", "# Outside\n")
	writeRel(t, dir, "README.md", "# Home\n\n[outside](../outside.md)\n")

	page, err := BuildDirPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d", len(page.Documents))
	}
	if page.Documents[1].RelPath != "../outside.md" || !page.Documents[1].Linked {
		t.Fatalf("linked doc = %#v", page.Documents[1])
	}
	if len(page.Tree) != 1 || page.Tree[0].Name != "README.md" {
		t.Fatalf("tree = %#v", page.Tree)
	}

	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	if !strings.Contains(html, "链接文档") {
		t.Fatalf("linked group label missing: %s", html)
	}
	if !strings.Contains(html, `data-doc-target="doc-1"`) {
		t.Fatalf("linked document button missing: %s", html)
	}
}
