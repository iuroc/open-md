package render

import (
	"net/url"
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

func TestBuildFilePageLinkedDocumentHasDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("[Other](other.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if page.Documents[1].Dir != dir {
		t.Fatalf("linked dir = %q, want %q", page.Documents[1].Dir, dir)
	}
}

func TestCrossDocumentEnglishAnchor(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other.md")
	if err := os.WriteFile(other, []byte("# Other\n\n## Usage\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[用法](other.md#usage)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d, want 2", len(page.Documents))
	}
	if !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1&amp;anchor=usage`) {
		t.Fatalf("English anchor not rewritten: %s", page.Documents[0].HTML)
	}
	if !strings.Contains(string(page.Documents[1].HTML), `id="doc-1-usage"`) {
		t.Fatalf("target heading ID missing: %s", page.Documents[1].HTML)
	}
}

func TestCrossDocumentChineseAnchor(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "other.md")
	if err := os.WriteFile(other, []byte("# 中文标题\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[说明](other.md#中文标题)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d, want 2", len(page.Documents))
	}
	escaped := url.QueryEscape("中文标题")
	if !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1&amp;anchor=`+escaped) {
		t.Fatalf("Chinese anchor not rewritten: %s", page.Documents[0].HTML)
	}
	if !strings.Contains(string(page.Documents[1].HTML), `id="doc-1-中文标题"`) {
		t.Fatalf("target heading ID missing: %s", page.Documents[1].HTML)
	}
}

func TestFootnoteIDsStayInEachDocument(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n\ntext[^1]\n\n[^1]: note a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("# B\n\ntext[^1]\n\n[^1]: note b\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildDirPage(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d, want 2", len(page.Documents))
	}
	if !strings.Contains(string(page.Documents[0].HTML), `id="doc-0-fn:1"`) ||
		!strings.Contains(string(page.Documents[1].HTML), `id="doc-1-fn:1"`) {
		t.Fatalf("footnote IDs not scoped: %s\n%s", page.Documents[0].HTML, page.Documents[1].HTML)
	}
	if !strings.Contains(string(page.Documents[1].HTML), `href="#doc-1-fnref:1"`) {
		t.Fatalf("backlink does not target its own document: %s", page.Documents[1].HTML)
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(out), `id="fn:1"`) != 0 || strings.Count(string(out), `id="fnref:1"`) != 0 {
		t.Fatal("unscoped footnote IDs remain in rendered page")
	}
}

func TestBuildFilePagePreRendersLinkedDocumentWithPercentName(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "100%.md")
	if err := os.WriteFile(other, []byte("# Percent\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[Other](100%.md)\n"), 0644); err != nil {
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
		t.Fatal("percent-named linked document not marked linked")
	}
	if !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1`) {
		t.Fatalf("link not rewritten: %s", page.Documents[0].HTML)
	}
}

func TestKnownDocumentWithPercentName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "100%.md"), []byte("# Percent\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[Other](100%.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 || !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1`) {
		t.Fatalf("percent-named document was not linked: docs=%d html=%s", len(page.Documents), page.Documents[0].HTML)
	}
}

func TestKnownDocumentWithEncodedSpace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a b.md"), []byte("# Spaced\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[Other](a%20b.md)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 || !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1`) {
		t.Fatalf("encoded-space document was not linked: docs=%d html=%s", len(page.Documents), page.Documents[0].HTML)
	}
}

func TestCrossDocumentAnchorWithPercentName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "100%.md"), []byte("# Percent\n\n## 用法\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "main.md")
	if err := os.WriteFile(main, []byte("# Main\n\n[说明](100%.md#用法)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	page, err := BuildFilePage(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("docs = %d, want 2", len(page.Documents))
	}
	escaped := url.QueryEscape("用法")
	if !strings.Contains(string(page.Documents[0].HTML), `#doc=doc-1&amp;anchor=`+escaped) {
		t.Fatalf("percent-name anchor not rewritten: %s", page.Documents[0].HTML)
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
