package render

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSiteFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildSite(t *testing.T, dir, cache string) *Site {
	t.Helper()
	site, err := BuildDirSite(dir, SiteOptions{
		CacheBase:  cache,
		Version:    "v1",
		AllowCache: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return site
}

func TestDocPageEnablesViewTransitionsInline(t *testing.T) {
	for _, want := range []string{
		"@view-transition { navigation: auto; }",
		"::view-transition { pointer-events: none; }",
		`addEventListener("pageswap"`,
		`addEventListener("pagereveal"`,
		"skipTransition()",
	} {
		if !strings.Contains(docPageTemplateSource, want) {
			t.Fatalf("doc page missing %q", want)
		}
	}
	if strings.Contains(readerCSS, "@view-transition") {
		t.Fatal("view transition opt-in must stay inline, not in an external stylesheet")
	}
}

func TestDocPagePreloadsOwnPayload(t *testing.T) {
	for _, want := range []string{
		`data-own-payload="{{.Payload}}"`,
		`<script src="{{.Payload}}"></script>`,
		"window.OPENMD_PAGES",
		"openmd-hash-load",
		"openmd-booting",
		"openmd-ready",
	} {
		if !strings.Contains(docPageTemplateSource, want) {
			t.Fatalf("doc page missing %q", want)
		}
	}
}

func TestBuildDirSiteCreatesIndependentPages(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "README.md", "# Home\n\n[B](sub/b.md)\n")
	bPath := writeSiteFile(t, dir, "sub/b.md", "# B\n\n## Usage\n")

	site := buildSite(t, dir, cache)
	if site.Generated != 2 {
		t.Fatalf("generated = %d", site.Generated)
	}
	if filepath.Base(site.InitialPage) == "" {
		t.Fatal("initial page missing")
	}
	if !fileExists(site.InitialPage) {
		t.Fatal("initial page was not written")
	}
	page, err := os.ReadFile(site.InitialPage)
	if err != nil {
		t.Fatal(err)
	}
	href := docHref(bPath)
	if !strings.Contains(string(page), `href="reader.css"`) {
		t.Fatal("shared reader.css not referenced")
	}
	if !strings.Contains(string(page), `name="color-scheme" content="light dark"`) {
		t.Fatal("early color-scheme meta missing")
	}
	if !strings.Contains(string(page), "html { background: #fbfdff; }") {
		t.Fatal("boot background missing")
	}
	if !strings.Contains(string(page), `class="doc-section`) {
		t.Fatal("initial page must embed its section for first paint")
	}
	bPage, err := os.ReadFile(filepath.Join(filepath.Dir(site.InitialPage), href))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(bPage), `class="doc-section`) {
		t.Fatal("non-initial pages must stay thin")
	}
	if !strings.Contains(string(page), `data-doc-href="`) {
		t.Fatal("thin document shell missing payload href")
	}
	payloadName := strings.TrimSuffix(filepath.Base(site.InitialPage), ".html") + ".js"
	if !strings.Contains(string(page), `data-own-payload="`+payloadName+`"`) {
		t.Fatal("thin document shell missing own payload")
	}
	if !strings.Contains(string(page), `<script src="`+payloadName+`"></script>`) {
		t.Fatal("own payload must load synchronously")
	}
	if !strings.Contains(string(page), "window.OPENMD_PAGES") {
		t.Fatal("inline section bootstrap missing")
	}
	for _, want := range []string{
		`id="mode-button"`,
		`id="mode-menu"`,
		`data-mode="split"`,
		`role="menuitemradio"`,
		`id="source-wrap-toggle"`,
		`role="switch"`,
	} {
		if !strings.Contains(string(page), want) {
			t.Fatalf("view mode menu missing %q", want)
		}
	}
	payload, err := os.ReadFile(strings.TrimSuffix(site.InitialPage, ".html") + ".js")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "window.OPENMD_PAGES") ||
		!strings.Contains(string(payload), `"sectionHTML"`) ||
		!strings.Contains(string(payload), "doc-section") {
		t.Fatalf("document payload missing section: %s", payload[:min(len(payload), 200)])
	}
	if strings.Contains(string(payload), "<!doctype") || strings.Contains(string(payload), `id="sidebar"`) {
		t.Fatal("document payload must not include the page shell")
	}
	if strings.Contains(string(payload), `"dir":""`) {
		t.Fatal("document payload directory missing")
	}
	if !strings.Contains(string(page), `src="tree.js"`) {
		t.Fatal("shared tree.js not referenced")
	}
	treeJS, err := os.ReadFile(filepath.Join(filepath.Dir(site.InitialPage), "tree.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(treeJS), "sub") || !strings.Contains(string(treeJS), href) {
		t.Fatalf("tree.js missing navigation: %s", treeJS)
	}
	if !strings.Contains(string(treeJS), "draggable=") {
		t.Fatal("tree links must not be draggable")
	}
}

func TestBuildDirSiteReusesUnchangedPages(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	writeSiteFile(t, dir, "b.md", "# B\n")

	first := buildSite(t, dir, cache)
	if first.Generated != 2 {
		t.Fatalf("first generated = %d", first.Generated)
	}
	second := buildSite(t, dir, cache)
	if second.Generated != 0 || second.Reused != 2 {
		t.Fatalf("second generated=%d reused=%d", second.Generated, second.Reused)
	}
}

func TestBuildDirSiteRebuildsOnlyChangedDocument(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	bPath := writeSiteFile(t, dir, "b.md", "# B\n")
	buildSite(t, dir, cache)

	if err := os.WriteFile(bPath, []byte("# B changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	second := buildSite(t, dir, cache)
	if second.Generated != 1 || second.Reused != 1 {
		t.Fatalf("generated=%d reused=%d", second.Generated, second.Reused)
	}
}

func TestBuildDirSiteDocumentSetChangeRebuildsAndPrunes(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	bPath := writeSiteFile(t, dir, "b.md", "# B\n")
	first := buildSite(t, dir, cache)
	bHref := docHref(bPath)

	cPath := writeSiteFile(t, dir, "c.md", "# C\n")
	_ = cPath
	second := buildSite(t, dir, cache)
	if second.Generated != 3 {
		t.Fatalf("generated = %d, want 3", second.Generated)
	}

	if err := os.Remove(bPath); err != nil {
		t.Fatal(err)
	}
	third := buildSite(t, dir, cache)
	if third.Generated != 2 {
		t.Fatalf("generated after delete = %d, want 2", third.Generated)
	}
	if fileExists(filepath.Join(filepath.Dir(first.InitialPage), bHref)) {
		t.Fatal("deleted document page was not pruned")
	}
}

func TestBuildDirSiteVersionChangeRebuilds(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	if _, err := BuildDirSite(dir, SiteOptions{CacheBase: cache, Version: "v1", AllowCache: true}); err != nil {
		t.Fatal(err)
	}
	site, err := BuildDirSite(dir, SiteOptions{CacheBase: cache, Version: "v2", AllowCache: true})
	if err != nil {
		t.Fatal(err)
	}
	if site.Generated != 1 {
		t.Fatalf("generated = %d", site.Generated)
	}
}

func TestBuildDirSiteConfirmationCanCancel(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	writeSiteFile(t, dir, "b.md", "# B\n")
	called := false
	_, err := BuildDirSite(dir, SiteOptions{
		CacheBase:         cache,
		Version:           "v1",
		AllowCache:        true,
		FullThresholdDocs: 1,
		Confirm: func(estimate Estimate) LoadChoice {
			called = true
			if estimate.PendingDocs != 2 {
				t.Fatalf("pending docs = %d", estimate.PendingDocs)
			}
			return LoadCancel
		},
	})
	if !errors.Is(err, ErrLoadCancelled) {
		t.Fatalf("err = %v", err)
	}
	if !called {
		t.Fatal("confirmation was not called")
	}
}

func TestBuildDirSiteCachedOpenDoesNotAsk(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "a.md", "# A\n")
	buildSite(t, dir, cache)

	called := false
	site, err := BuildDirSite(dir, SiteOptions{
		CacheBase:         cache,
		Version:           "v1",
		AllowCache:        true,
		FullThresholdDocs: 1,
		Confirm: func(estimate Estimate) LoadChoice {
			called = true
			return LoadCancel
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("cached open should not ask for confirmation")
	}
	if site.Reused != 1 {
		t.Fatalf("reused = %d", site.Reused)
	}
}

func TestBuildFileSiteAsksBeforeOversizedFile(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	path := filepath.Join(dir, "big.md")
	if err := os.WriteFile(path, []byte("# Big\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, 33<<20); err != nil {
		t.Fatal(err)
	}
	called := false
	_, err := BuildFileSite(path, SiteOptions{
		CacheBase:  cache,
		Version:    "v1",
		AllowCache: true,
		Confirm: func(estimate Estimate) LoadChoice {
			called = true
			return LoadCancel
		},
	})
	if !errors.Is(err, ErrLoadCancelled) {
		t.Fatalf("err = %v", err)
	}
	if !called {
		t.Fatal("confirmation was not called")
	}
}

func TestBuildFileSiteCreatesLinkedPage(t *testing.T) {
	dir := t.TempDir()
	cache := t.TempDir()
	writeSiteFile(t, dir, "other.md", "# Other\n")
	mainPath := writeSiteFile(t, dir, "main.md", "# Main\n\n[Other](other.md)\n")

	site, err := BuildFileSite(mainPath, SiteOptions{CacheBase: cache, Version: "v1", AllowCache: true})
	if err != nil {
		t.Fatal(err)
	}
	if site.Generated != 2 {
		t.Fatalf("generated = %d", site.Generated)
	}
	treeJS, err := os.ReadFile(filepath.Join(filepath.Dir(site.InitialPage), "tree.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(treeJS), "链接文档") {
		t.Fatalf("linked document group missing: %s", treeJS)
	}
}
