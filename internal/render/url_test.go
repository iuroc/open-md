package render

import (
	"strings"
	"testing"
)

func TestRewriteLocalImagePaths(t *testing.T) {
	base := `C:\docs with space\notes`
	got, err := RewriteURL("images/a b.png", base, ImageURL, nil)
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
	got, err := RewriteURL("图片/图 1.png", `C:\文档`, ImageURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "%E5%9B%BE%E7%89%87") || !strings.Contains(got, "%E5%9B%BE%201.png") {
		t.Fatalf("unicode encoding missing: %q", got)
	}
	unc, err := RewriteURL(`\\server\share\a.png`, `C:\docs`, ImageURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if unc != "file://server/share/a.png" {
		t.Fatalf("unc = %q", unc)
	}
}

func TestRewriteRemoteAnchorAndDataURLsAreUnchanged(t *testing.T) {
	for _, raw := range []string{"https://example.com/a.png", "#section", "data:image/png;base64,AAAA"} {
		got, err := RewriteURL(raw, `C:\docs`, ImageURL, nil)
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
		got, err := RewriteURL(raw, `C:\docs`, LinkURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	}
}

func TestRewriteUnknownSchemeIsBlocked(t *testing.T) {
	got, err := RewriteURL("ms-msdt:/id PCWDiagnostic", `C:\docs`, LinkURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestRewriteKnownDocumentLink(t *testing.T) {
	known := map[string]string{`c:\docs\other.md`: "doc-1"}
	got, err := RewriteURL("other.md#usage", `C:\docs`, LinkURL, known)
	if err != nil {
		t.Fatal(err)
	}
	if got != "#doc=doc-1&anchor=usage" {
		t.Fatalf("got %q", got)
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
