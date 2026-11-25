package manual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFingerprintChangesWithContent(t *testing.T) {
	a := fstest.MapFS{"index.md": {Data: []byte("# A\n")}}
	b := fstest.MapFS{"index.md": {Data: []byte("# B\n")}}
	ha, err := fingerprint(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := fingerprint(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha == "" || ha == hb {
		t.Fatalf("fingerprints must differ: %q %q", ha, hb)
	}
}

func TestExtractWritesCompleteMarker(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "src-abc")
	fsys := fstest.MapFS{"index.md": {Data: []byte("# 测试手册\n")}}
	if err := extract(fsys, dst); err != nil {
		t.Fatal(err)
	}
	if !complete(dst) {
		t.Fatal("missing .complete marker")
	}
	data, err := os.ReadFile(filepath.Join(dst, "index.md"))
	if err != nil || !strings.Contains(string(data), "测试手册") {
		t.Fatalf("chapter not written: %v %q", err, data)
	}
}

func TestExtractReplacesIncompleteDirectory(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "src-abc")
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "index.md"), []byte("# old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fsys := fstest.MapFS{"index.md": {Data: []byte("# new\n")}}
	if err := extract(fsys, dst); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dst, "index.md"))
	if !strings.Contains(string(data), "new") {
		t.Fatalf("incomplete directory was reused: %q", data)
	}
	if !complete(dst) {
		t.Fatal("missing .complete marker")
	}
}
