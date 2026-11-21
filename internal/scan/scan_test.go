package scan

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, rel string, data []byte) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func relPaths(result *Result) []string {
	paths := make([]string, 0, len(result.Documents))
	for _, doc := range result.Documents {
		paths = append(paths, doc.RelPath)
	}
	return paths
}

func TestDirectoryScansRecursively(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b.MD", []byte("b"))
	write(t, dir, "a.markdown", []byte("a"))
	write(t, dir, "note.txt", []byte("x"))
	write(t, dir, "sub/nested.md", []byte("n"))

	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.markdown", "b.MD", "sub/nested.md"}
	if strings.Join(relPaths(got), "|") != strings.Join(want, "|") {
		t.Fatalf("rel paths = %#v, want %#v", relPaths(got), want)
	}
}

func TestDirectoryLoadsNestedDocumentsWithoutRootMarkdown(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "specs/source-view/spec.md", []byte("# source-view"))
	write(t, dir, "changes/archive/foo/proposal.md", []byte("# proposal"))

	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"specs/source-view/spec.md", "changes/archive/foo/proposal.md"}
	if strings.Join(relPaths(got), "|") != strings.Join(want, "|") {
		t.Fatalf("rel paths = %#v, want %#v", relPaths(got), want)
	}
}

func TestDirectoryNaturalSort(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"file10.md", "file2.md", "file1.md"} {
		write(t, dir, name, []byte("x"))
	}
	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"file1.md", "file2.md", "file10.md"}
	if strings.Join(relPaths(got), "|") != strings.Join(want, "|") {
		t.Fatalf("rel paths = %#v, want %#v", relPaths(got), want)
	}
}

func TestDirectoryLimits(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		write(t, dir, name, []byte("12345"))
	}
	got, err := Directory(dir, Limits{MaxDocs: 2, MaxTotalBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 2 || !got.Truncated {
		t.Fatalf("docs=%d truncated=%v", len(got.Documents), got.Truncated)
	}
}

func TestDirectorySkipsOversizedFileBeforeSmall(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a-large.md", make([]byte, 20))
	write(t, dir, "b-small.md", []byte("12345"))
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxFileBytes: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].Name != "b-small.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
	if len(got.Skipped) != 1 || got.Skipped[0].Name != "a-large.md" {
		t.Fatalf("skipped = %#v", got.Skipped)
	}
	if len(got.Warnings) == 0 || !strings.Contains(got.Warnings[0], "a-large.md") {
		t.Fatalf("warnings = %#v", got.Warnings)
	}
}

func TestDirectoryNestedOversizedWarningUsesRelativePath(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "sub/big.md", make([]byte, 20))
	write(t, dir, "sub/small.md", []byte("12345"))
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxFileBytes: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].RelPath != "sub/small.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "sub/big.md") {
		t.Fatalf("warnings = %#v", got.Warnings)
	}
}

func TestDirectorySkipsOversizedFileAfterSmall(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a-small.md", []byte("12345"))
	write(t, dir, "b-large.md", make([]byte, 20))
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxFileBytes: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].Name != "a-small.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
	if len(got.Skipped) != 1 || got.Skipped[0].Name != "b-large.md" {
		t.Fatalf("skipped = %#v", got.Skipped)
	}
}

func TestDirectoryTotalBytesLimit(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.md", []byte("12345"))
	write(t, dir, "b.md", []byte("12345"))
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || !got.Truncated {
		t.Fatalf("docs=%d truncated=%v", len(got.Documents), got.Truncated)
	}
}

func TestDirectoryAllOversized(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.md", make([]byte, 20))
	write(t, dir, "sub/b.md", make([]byte, 20))
	_, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxFileBytes: 10})
	if !errors.Is(err, ErrNoLoadableMarkdown) {
		t.Fatalf("err = %v, want ErrNoLoadableMarkdown", err)
	}
	if !strings.Contains(err.Error(), "大小限制") {
		t.Fatalf("error is not understandable: %v", err)
	}
}

func TestDirectoryEmpty(t *testing.T) {
	_, err := Directory(t.TempDir(), DefaultLimits)
	if !errors.Is(err, ErrNoMarkdown) {
		t.Fatalf("err = %v", err)
	}
}

func TestDirectorySkipsIgnoredDirectories(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".git/a.md", []byte("a"))
	write(t, dir, "node_modules/b.md", []byte("b"))
	write(t, dir, "vendor/c.md", []byte("c"))
	write(t, dir, "docs/keep.md", []byte("k"))

	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].RelPath != "docs/keep.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
}

func TestDirectoryDoesNotFollowSymlinks(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "real/a.md", []byte("a"))
	link := filepath.Join(dir, "link")
	if err := os.Symlink(filepath.Join(dir, "real"), link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].RelPath != "real/a.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
}

func TestDirectoryDoesNotFollowJunctions(t *testing.T) {
	if _, err := exec.LookPath("cmd"); err != nil {
		t.Skip("cmd unavailable")
	}
	dir := t.TempDir()
	write(t, dir, "real/a.md", []byte("a"))
	junction := filepath.Join(dir, "junction")
	if err := exec.Command("cmd", "/c", "mklink", "/J", junction, filepath.Join(dir, "real")).Run(); err != nil {
		t.Skipf("junction unavailable: %v", err)
	}
	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || got.Documents[0].RelPath != "real/a.md" {
		t.Fatalf("docs = %#v", got.Documents)
	}
}

func TestDirectoryDepthLimit(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "top.md", []byte("t"))
	write(t, dir, "sub/a.md", []byte("a"))
	write(t, dir, "sub/deep/b.md", []byte("b"))
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxDepth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 2 {
		t.Fatalf("docs = %#v", relPaths(got))
	}
	if !hasWarning(got.Warnings, "深度") {
		t.Fatalf("warnings = %#v", got.Warnings)
	}
}

func TestDirectoryEntryBudget(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		write(t, dir, name, []byte("x"))
	}
	got, err := Directory(dir, Limits{MaxDocs: 10, MaxTotalBytes: 100, MaxEntries: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 2 {
		t.Fatalf("docs = %#v", relPaths(got))
	}
	if !hasWarning(got.Warnings, "目录项") {
		t.Fatalf("warnings = %#v", got.Warnings)
	}
}

func TestDirectoryTruncationPrefersShallowDocuments(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "z.md", []byte("z"))
	write(t, dir, "a/a.md", []byte("a"))
	write(t, dir, "b/c/b.md", []byte("b"))
	got, err := Directory(dir, Limits{MaxDocs: 2, MaxTotalBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"z.md", "a/a.md"}
	if strings.Join(relPaths(got), "|") != strings.Join(want, "|") {
		t.Fatalf("rel paths = %#v, want %#v", relPaths(got), want)
	}
	if !got.Truncated {
		t.Fatal("expected truncation")
	}
}

func TestDirectoryRescanIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b/2.md", []byte("2"))
	write(t, dir, "b/10.md", []byte("10"))
	write(t, dir, "a.md", []byte("a"))

	first, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(relPaths(first), "|") != strings.Join(relPaths(second), "|") {
		t.Fatalf("first = %#v, second = %#v", relPaths(first), relPaths(second))
	}
}

func TestDirectoryUnlimitedIgnoresCaps(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 600; i++ {
		write(t, dir, filepath.Join("group", fmt.Sprintf("doc-%04d.md", i)), []byte("x"))
	}
	got, err := Directory(dir, UnlimitedLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 600 || got.Truncated {
		t.Fatalf("docs=%d truncated=%v", len(got.Documents), got.Truncated)
	}
}

func TestDirectoryUnlimitedDoesNotSkipLargeFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "big.md", make([]byte, 20))
	got, err := Directory(dir, UnlimitedLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 1 || len(got.Skipped) != 0 {
		t.Fatalf("docs=%#v skipped=%#v", got.Documents, got.Skipped)
	}
}

func hasWarning(warnings []string, want string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, want) {
			return true
		}
	}
	return false
}
