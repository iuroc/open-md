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
