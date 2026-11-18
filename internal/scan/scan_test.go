package scan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectoryScansCurrentLevelOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.MD"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.markdown"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "nested.md"), []byte("n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Documents) != 2 {
		t.Fatalf("docs = %#v", got.Documents)
	}
	if got.Documents[0].Name != "a.markdown" || got.Documents[1].Name != "b.MD" {
		t.Fatalf("order = %#v", got.Documents)
	}
}

func TestDirectoryNaturalSort(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"file10.md", "file2.md", "file1.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Directory(dir, DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{got.Documents[0].Name, got.Documents[1].Name, got.Documents[2].Name}
	want := []string{"file1.md", "file2.md", "file10.md"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %#v", names)
		}
	}
}

func TestDirectoryLimits(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("12345"), 0644); err != nil {
			t.Fatal(err)
		}
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
	if err := os.WriteFile(filepath.Join(dir, "a-large.md"), make([]byte, 20), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b-small.md"), []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}
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

func TestDirectorySkipsOversizedFileAfterSmall(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a-small.md"), []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b-large.md"), make([]byte, 20), 0644); err != nil {
		t.Fatal(err)
	}
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
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}
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
	for _, name := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, 20), 0644); err != nil {
			t.Fatal(err)
		}
	}
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
