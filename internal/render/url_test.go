package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteLocalImagePaths(t *testing.T) {
	base := `C:\docs with space\notes`
	got, err := RewriteURL("images/a b.png", base, ImageURL, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "file:///C:/docs%20with%20space/notes/images/") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, "a%20b.png") {
		t.Fatalf("space encoding missing: %q", got)
	}
}

func TestRewriteChinesePathAndUNC(t *testing.T) {
	got, err := RewriteURL("图片/图 1.png", `C:\文档`, ImageURL, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "%E5%9B%BE%E7%89%87") || !strings.Contains(got, "%E5%9B%BE%201.png") {
		t.Fatalf("unicode encoding missing: %q", got)
	}
	unc, err := RewriteURL(`\\server\share\a.png`, `C:\docs`, ImageURL, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if unc != "file://server/share/a.png" {
		t.Fatalf("unc = %q", unc)
	}
}

func TestRewriteRemoteAnchorAndDataURLsAreUnchanged(t *testing.T) {
	for _, raw := range []string{"https://example.com/a.png", "#section", "data:image/png;base64,AAAA"} {
		got, err := RewriteURL(raw, `C:\docs`, ImageURL, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != raw {
			t.Fatalf("raw %q became %q", raw, got)
		}
	}
}

func TestRewriteDangerousSchemes(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "vbscript:msgbox(1)"} {
		got, err := RewriteURL(raw, `C:\docs`, LinkURL, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	}
}

func TestRewriteUnknownSchemeIsBlocked(t *testing.T) {
	got, err := RewriteURL("ms-msdt:/id PCWDiagnostic", `C:\docs`, LinkURL, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestRewriteKnownDocumentLink(t *testing.T) {
	known := map[string]string{`c:\docs\other.md`: "doc-1"}
	got, err := RewriteURL("other.md#usage", `C:\docs`, LinkURL, known, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "#doc=doc-1&anchor=usage" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveLocalDestinationLiteralFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"100%.md", "foo%bar.md", "a b.md", "a#b.md", "other.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# x\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		raw      string
		wantPath string
		wantFrag string
		wantQ    string
	}{
		{"100%.md", filepath.Join(dir, "100%.md"), "", ""},
		{"foo%bar.md", filepath.Join(dir, "foo%bar.md"), "", ""},
		{"a%20b.md", filepath.Join(dir, "a b.md"), "", ""},
		{"a#b.md", filepath.Join(dir, "a#b.md"), "", ""},
		{"other.md#usage", filepath.Join(dir, "other.md"), "usage", ""},
		{"image.png?raw=1", filepath.Join(dir, "image.png"), "", "raw=1"},
		{"100%missing.md", filepath.Join(dir, "100%missing.md"), "", ""},
	}
	for _, tc := range cases {
		target, ok := resolveLocalDestination(tc.raw, dir)
		if !ok {
			t.Fatalf("%q: not resolved", tc.raw)
		}
		if target.Path != tc.wantPath || target.Fragment != tc.wantFrag || target.RawQuery != tc.wantQ {
			t.Fatalf("%q = %#v, want path=%q fragment=%q query=%q", tc.raw, target, tc.wantPath, tc.wantFrag, tc.wantQ)
		}
	}
}

func TestResolveLocalDestinationPrefersLiteralPath(t *testing.T) {
	dir := t.TempDir()
	literal := filepath.Join(dir, "a%20b.md")
	if err := os.WriteFile(literal, []byte("# literal\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a b.md"), []byte("# spaced\n"), 0644); err != nil {
		t.Fatal(err)
	}

	target, ok := resolveLocalDestination("a%20b.md", dir)
	if !ok {
		t.Fatal("not resolved")
	}
	if target.Path != literal {
		t.Fatalf("path = %q, want literal %q", target.Path, literal)
	}
}

func TestResolveLocalDestinationRejectsRemoteSchemes(t *testing.T) {
	if _, ok := resolveLocalDestination("https://example.com/a.md", `C:\docs`); ok {
		t.Fatal("https target must not resolve as a local path")
	}
	if _, ok := resolveLocalDestination("mailto:a@example.com", `C:\docs`); ok {
		t.Fatal("mailto target must not resolve as a local path")
	}
}

func TestMalformedLocalDestinationIsDegraded(t *testing.T) {
	got, err := RewriteURL("100%.md", `C:\docs`, LinkURL, nil, "")
	if err != nil {
		t.Fatalf("malformed destination returned error: %v", err)
	}
	if got == "" {
		t.Fatal("malformed destination should degrade to a file URL, not disappear")
	}
}

func TestMalformedLinkKeepsDocumentReadable(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# Title\n\n[bad](100%.md)\n\n![bad](100%.png)\n\n正文\n"), `C:\docs\note.md`, RenderOptions{IDPrefix: "doc-0"})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, "正文") {
		t.Fatalf("document body disappeared: %s", html)
	}
	if strings.Contains(html, "无法读取") {
		t.Fatalf("document was replaced by an error page: %s", html)
	}
}

func TestRenderMarkdownRewritesSameDocumentAnchor(t *testing.T) {
	doc, err := RenderMarkdown([]byte("# Title\n\n[jump](#install)\n\n## Install\n"), `C:\docs\note.md`, RenderOptions{
		IDPrefix:     "doc-0",
		CurrentDocID: "doc-0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc.HTML), `href="#doc=doc-0&amp;anchor=install"`) {
		t.Fatalf("same-document anchor not rewritten: %s", doc.HTML)
	}
}

func TestRenderMarkdownRewritesImageDestination(t *testing.T) {
	doc, err := RenderMarkdown([]byte("![x](pic.png)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc.HTML), "file:///C:/docs/pic.png") {
		t.Fatalf("image URL not rewritten: %s", doc.HTML)
	}
}

func TestRenderMarkdownKeepsEncodedImageURL(t *testing.T) {
	doc, err := RenderMarkdown([]byte("![x](<images/图 1.png>)\n"), `C:\文档\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, "file:///C:/%E6%96%87%E6%A1%A3/images/%E5%9B%BE%201.png") {
		t.Fatalf("encoded image URL missing: %s", html)
	}
}
