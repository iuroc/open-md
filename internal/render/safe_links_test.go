package render

import (
	"strings"
	"testing"
)

func TestExternalLinkHTTPOpensInNewTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[示例](https://example.com)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, `target="_blank"`) || !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("external link missing new-tab attributes: %s", html)
	}
}

func TestExternalLinkFileOpensInNewTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[手册](manual.pdf)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, `target="_blank"`) || !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("local file link missing new-tab attributes: %s", html)
	}
}

func TestProtocolRelativeExternalLinkOpensInNewTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[示例](//example.com/a)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, `target="_blank"`) || !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("protocol-relative link missing new-tab attributes: %s", html)
	}
}

func TestBareURLAutolinkOpensInNewTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("访问 https://example.com 了解详情\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if !strings.Contains(html, `target="_blank"`) || !strings.Contains(html, `rel="noopener noreferrer"`) {
		t.Fatalf("bare URL autolink missing new-tab attributes: %s", html)
	}
}

func TestBareEmailAutolinkStaysInCurrentTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("联系 a@example.com\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if strings.Contains(html, `target="_blank"`) {
		t.Fatalf("email autolink must not force a new tab: %s", html)
	}
	if !strings.Contains(html, `href="mailto:a@example.com"`) {
		t.Fatalf("email autolink missing mailto href: %s", html)
	}
}

func TestInternalLinkStaysInCurrentTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[跳到安装](#安装)\n"), `C:\docs\note.md`, RenderOptions{
		IDPrefix:     "doc-0",
		CurrentDocID: "doc-0",
	})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if strings.Contains(html, `target="_blank"`) {
		t.Fatalf("internal anchor must stay in current tab: %s", html)
	}
	if !strings.Contains(html, `#doc=doc-0&amp;anchor=`) {
		t.Fatalf("internal anchor not rewritten: %s", html)
	}
}

func TestMailtoLinkStaysInCurrentTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[邮件](mailto:a@example.com)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if strings.Contains(html, `target="_blank"`) {
		t.Fatalf("mailto link must not force a new tab: %s", html)
	}
}

func TestTelLinkStaysInCurrentTab(t *testing.T) {
	doc, err := RenderMarkdown([]byte("[电话](tel:123456)\n"), `C:\docs\note.md`, RenderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	html := string(doc.HTML)
	if strings.Contains(html, `target="_blank"`) {
		t.Fatalf("tel link must not force a new tab: %s", html)
	}
}
