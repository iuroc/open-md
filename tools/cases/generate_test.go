package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func SmallTestSizes() BigSizes {
	return BigSizes{Oversize: 64 << 10, TotalFile: 16 << 10, AtLimit: 64 << 10}
}

func TestPrepareRefusesUnmarkedDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("user data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := prepareOutputDir(dir, false); err == nil {
		t.Fatal("expected refusal for unmarked non-empty dir")
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); err != nil {
		t.Fatal("user file must survive refusal")
	}
}

func TestPrepareForceClearsUnmarkedDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("user data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := prepareOutputDir(dir, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); !os.IsNotExist(err) {
		t.Fatal("force should clear the dir")
	}
}

func TestGenerateCleansMarkedDir(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(Options{Out: dir, Sizes: SmallTestSizes()}); err != nil {
		t.Fatal(err)
	}
	stray := filepath.Join(dir, "stray.md")
	if err := os.WriteFile(stray, []byte("stray"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(Options{Out: dir, Sizes: SmallTestSizes()}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stray); !os.IsNotExist(err) {
		t.Fatal("marked dir should be cleared before regeneration")
	}
	if _, err := os.Stat(filepath.Join(dir, markerName)); err != nil {
		t.Fatal("marker must exist after generation")
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	if err := Generate(Options{Out: first, Sizes: SmallTestSizes()}); err != nil {
		t.Fatal(err)
	}
	if err := Generate(Options{Out: second, Sizes: SmallTestSizes()}); err != nil {
		t.Fatal(err)
	}
	compareTrees(t, first, second)
}

func compareTrees(t *testing.T, a, b string) {
	t.Helper()
	filesA := collectFiles(t, a)
	filesB := collectFiles(t, b)
	if !slices.Equal(filesA, filesB) {
		t.Fatalf("file lists differ:\nfirst:  %v\nsecond: %v", filesA, filesB)
	}
	for _, rel := range filesA {
		dataA, err := os.ReadFile(filepath.Join(a, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		dataB, err := os.ReadFile(filepath.Join(b, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(dataA, dataB) {
			t.Fatalf("content differs for %s", rel)
		}
	}
}

func collectFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(files)
	return files
}
