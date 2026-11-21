package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2/lexers"
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

func TestLightFixtureStructure(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(Options{Out: dir}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"CHECKLIST.md", markerName,
		"01-headings.md", "05-code.md",
		"05a-code-systems.md", "05b-code-scripting.md",
		"05c-code-functional.md", "05d-code-data-web.md",
		"06-footnotes.md",
		"07-footnotes-peer.md", "08-anchors.md", "09-links-target.md",
		"10-frontmatter-yaml.md", "11-frontmatter-toml.md", "12-hr-mid.md",
		"13-links.md", "14-images.md", "15-broken.md", "16-empty.md",
		"17-bom.md", "18-unicode.md", "19-html.md", "20-long-line.md",
		"100%.md", "foo%bar.md", "a b.md", "中文文件名.md",
		"UPPER.MARKDOWN", "hash#name.md",
		"fallback/link.md", "fallback/a b.md",
		"literal/link.md", "literal/a%20b.md", "literal/a b.md",
		"sub/nested.md",
		"tree/root.md", "tree/guides/deep/deeper/leaf.md",
		"tree/中文目录/子目录/更多.md",
		"tree/no-md/note.txt", "tree/mixed/logo.png",
		"tree/duplicates/a.md", "tree/duplicates/b/a.md",
		"tree/sort-order/10.md",
		"limits-count/doc-001.md", "limits-count/doc-501.md",
		"common/pixel.png", "common/手册.pdf",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(want))); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	for _, absent := range []string{
		"limits/skip/big-33mib.md", "limits/total/01.md",
		"limits/all-big/big.md", "limits/at-limit/big-32mib.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(absent))); !os.IsNotExist(err) {
			t.Errorf("%s must not exist without -big", absent)
		}
	}
	info, err := os.Stat(filepath.Join(dir, "tree", "empty"))
	if err != nil || !info.IsDir() {
		t.Fatal("tree/empty must be an empty directory")
	}
	if size := mustSize(t, filepath.Join(dir, "16-empty.md")); size != 0 {
		t.Fatalf("16-empty.md must be empty, got %d", size)
	}
	if !bytes.HasPrefix(mustRead(t, filepath.Join(dir, "17-bom.md")), []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("17-bom.md must start with UTF-8 BOM")
	}
}

func TestLightFixtureContent(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(Options{Out: dir}); err != nil {
		t.Fatal(err)
	}
	yaml := string(mustRead(t, filepath.Join(dir, "10-frontmatter-yaml.md")))
	if !bytes.HasPrefix([]byte(yaml), []byte("---\n")) || !bytes.Contains([]byte(yaml), []byte("# 正文标题")) {
		t.Fatalf("unexpected YAML front matter fixture: %q", yaml[:min(len(yaml), 80)])
	}
	if !bytes.Contains(mustRead(t, filepath.Join(dir, "fallback", "link.md")), []byte("a%20b.md")) {
		t.Fatal("fallback/link.md must link a%20b.md")
	}
	if !bytes.Contains(mustRead(t, filepath.Join(dir, "15-broken.md")), []byte("101%.md")) {
		t.Fatal("15-broken.md must reference 101%.md")
	}
	if !bytes.Contains(mustRead(t, filepath.Join(dir, "14-images.md")), []byte("common/pixel.png")) {
		t.Fatal("14-images.md must reference common/pixel.png")
	}
	target := mustRead(t, filepath.Join(dir, "09-links-target.md"))
	if !bytes.Contains(target, []byte("## Usage")) || !bytes.Contains(target, []byte("# 中文标题")) {
		t.Fatal("09-links-target.md must expose Usage and 中文标题")
	}
	dupA := mustRead(t, filepath.Join(dir, "tree", "duplicates", "a.md"))
	dupB := mustRead(t, filepath.Join(dir, "tree", "duplicates", "b", "a.md"))
	if bytes.Equal(dupA, dupB) {
		t.Fatal("duplicate-name fixtures must differ")
	}
}

func TestCodeFixtureLanguages(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(Options{Out: dir}); err != nil {
		t.Fatal(err)
	}
	expected := map[string][]string{
		"05a-code-systems.md":    {"go", "rust", "c", "cpp", "csharp", "java", "kotlin", "swift", "scala", "nasm"},
		"05b-code-scripting.md":  {"python", "javascript", "typescript", "ruby", "php", "perl", "lua", "bash", "powershell", "r", "julia", "dart"},
		"05c-code-functional.md": {"haskell", "elixir", "erlang", "clojure", "fsharp", "ocaml", "scheme", "lisp"},
		"05d-code-data-web.md":   {"json", "yaml", "toml", "sql", "html", "css", "scss", "xml", "graphql", "protobuf", "ini", "diff", "dockerfile", "makefile", "nginx", "terraform", "latex"},
	}
	for file, languages := range expected {
		data := string(mustRead(t, filepath.Join(dir, file)))
		for _, lang := range languages {
			if !strings.Contains(data, "```"+lang+"\n") {
				t.Errorf("%s: missing fenced block for %q", file, lang)
			}
			if lexers.Get(lang) == nil {
				t.Errorf("%s: Chroma has no lexer for %q, block would not highlight", file, lang)
			}
		}
	}
}

func TestBigSizesDefaults(t *testing.T) {
	got := DefaultBigSizes()
	want := BigSizes{Oversize: 33 << 20, TotalFile: 11 << 19, AtLimit: 32 << 20}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestBigFixtureSizes(t *testing.T) {
	dir := t.TempDir()
	sizes := SmallTestSizes()
	if err := Generate(Options{Out: dir, Big: true, Sizes: sizes}); err != nil {
		t.Fatal(err)
	}
	checks := map[string]int64{
		"limits/skip/big-33mib.md":     sizes.Oversize,
		"limits/total/01.md":           sizes.TotalFile,
		"limits/total/05.md":           sizes.TotalFile,
		"limits/at-limit/big-32mib.md": sizes.AtLimit,
		"limits/all-big/big.md":        sizes.Oversize,
	}
	for rel, want := range checks {
		if got := mustSize(t, filepath.Join(dir, filepath.FromSlash(rel))); got != want {
			t.Errorf("%s: got %d want %d", rel, got, want)
		}
	}
	if !bytes.Equal(mustRead(t, filepath.Join(dir, "limits", "skip", "big-33mib.md")),
		mustRead(t, filepath.Join(dir, "limits", "all-big", "big.md"))) {
		t.Fatal("all-big must reuse the 33 MiB content")
	}
	if _, err := os.Stat(filepath.Join(dir, "limits", "skip", "small.md")); err != nil {
		t.Fatal("skip case needs a small neighbour document")
	}
}

func TestWriteRepeatedMarkdownExactSize(t *testing.T) {
	for _, size := range []int64{0, 1, 1023, 1 << 20} {
		var buf bytes.Buffer
		if err := writeRepeatedMarkdown(&buf, size); err != nil {
			t.Fatal(err)
		}
		if int64(buf.Len()) != size {
			t.Fatalf("size %d: wrote %d", size, buf.Len())
		}
		if !utf8.Valid(buf.Bytes()) {
			t.Fatalf("size %d: output is not valid UTF-8", size)
		}
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
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
