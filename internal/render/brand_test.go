package render

import (
	"html/template"
	"strings"
	"testing"
)

func assertWordmarkMarkup(t *testing.T, name, s string) {
	t.Helper()
	for _, want := range []string{
		`class="brand"`,
		`class="brand settings-title"`,
		`class="brand-mark"`,
		`viewBox="63 65 130 126"`,
		`points="78,176 78,80 128,132 178,80 178,176"`,
		`aria-hidden="true" focusable="false"`,
		`class="brand-wordmark"`,
		`class="brand-open">Open</span>`,
		`class="brand-md">MD</span>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("%s missing wordmark markup %q", name, want)
		}
	}
	if got := strings.Count(s, `class="brand-mark"`); got != 2 {
		t.Fatalf("%s brand marks = %d, want 2", name, got)
	}
	for _, banned := range []string{
		`<span>OpenMD</span>`,
		`class="settings-title">OpenMD</span>`,
	} {
		if strings.Contains(s, banned) {
			t.Fatalf("%s still contains the plain wordmark %q", name, banned)
		}
	}
}

func TestDocPageUsesWordmark(t *testing.T) {
	out, err := RenderDocPage(DocPageData{
		Title: "A",
		Doc:   Document{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertWordmarkMarkup(t, "docpage", string(out))
}

func TestPageUsesWordmark(t *testing.T) {
	page := &Page{
		Mode:  "file",
		Title: "A",
		Documents: []Document{{
			ID:    "doc-0",
			Title: "A",
			Path:  `C:\docs\a.md`,
			Dir:   `C:\docs`,
		}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	assertWordmarkMarkup(t, "page", string(out))
}
