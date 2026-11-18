package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildFilePagePreRendersLinkedLocalDocument(t *testing.T) {
	sourceDir := t.TempDir()
	otherDir := t.TempDir()
	other := filepath.Join(otherDir, "other.md")
	if err := os.WriteFile(other, []byte("# Other\n\n## Detail\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(sourceDir, other)
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(sourceDir, "main.md")
	if err := os.WriteFile(main, []byte("[Other](<"+filepath.ToSlash(rel)+">)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d, want 2", len(page.Documents))
	}
	if !page.Documents[1].Linked {
		t.Fatal("linked document not marked as linked")
	}
	if !strings.Contains(string(page.Documents[0].HTML), "#doc=doc-1") {
		t.Fatalf("link not rewritten: %s", page.Documents[0].HTML)
	}
	if !strings.Contains(string(page.Documents[1].HTML), `id="doc-1-detail"`) {
		t.Fatalf("linked document heading not prefixed: %s", page.Documents[1].HTML)
	}
}

func TestBuildFilePageSkipsNetworkLinkedDocument(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("[Net](file://server/share/other.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 1 {
		t.Fatalf("network document must not be preloaded, docs = %d", len(page.Documents))
	}
	if !strings.Contains(string(page.Documents[0].HTML), "file://server/share/other.md") {
		t.Fatalf("network link should remain a file URL: %s", page.Documents[0].HTML)
	}
}

func TestBuildFilePageDoesNotFollowLinkedDocumentsRecursively(t *testing.T) {
	dir := t.TempDir()
	third := filepath.Join(dir, "third.md")
	if err := os.WriteFile(third, []byte("# Third\n"), 0644); err != nil {
		t.Fatal(err)
	}
	second := filepath.Join(dir, "second.md")
	if err := os.WriteFile(second, []byte("# Second\n\n[Third](third.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(dir, "first.md")
	if err := os.WriteFile(first, []byte("# First\n\n[Second](second.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("linked docs must not be followed recursively, docs = %d", len(page.Documents))
	}
}

func TestBuildDirPagePreRendersLinkedDocumentOutsideDirectory(t *testing.T) {
	docsDir := t.TempDir()
	otherDir := t.TempDir()
	other := filepath.Join(otherDir, "external.md")
	if err := os.WriteFile(other, []byte("# External\n"), 0644); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(docsDir, other)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("[External](<"+filepath.ToSlash(rel)+">)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildDirPage(docsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 || !page.Documents[1].Linked {
		t.Fatalf("linked document missing: %#v", page.Documents)
	}
}

func TestIsNetworkPath(t *testing.T) {
	if !isNetworkPath(`\\server\share\a.md`) {
		t.Fatal("UNC path not detected as network")
	}
	if isNetworkPath(`C:\docs\a.md`) {
		t.Fatal("local drive path detected as network")
	}
}
