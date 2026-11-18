package render

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type safeLinkRenderer struct{}

func (r *safeLinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindImage, r.renderImage)
}

func (r *safeLinkRenderer) renderAutoLink(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	auto := n.(*ast.AutoLink)
	href := util.URLEscape(auto.URL(source), false)
	if auto.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(string(href)), "mailto:") {
		href = append([]byte("mailto:"), href...)
	}

	_, _ = w.WriteString(`<a href="`)
	if !gmhtml.IsDangerousURL(href) {
		_, _ = w.Write(util.EscapeHTML(href))
	}
	_ = w.WriteByte('"')
	if opensInNewTab(href) {
		_, _ = w.WriteString(` target="_blank" rel="noopener noreferrer"`)
	}
	if auto.Attributes() != nil {
		gmhtml.RenderAttributes(w, auto, gmhtml.LinkAttributeFilter)
	}
	_ = w.WriteByte('>')
	_, _ = w.Write(util.EscapeHTML(auto.Label(source)))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

func (r *safeLinkRenderer) renderLink(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}
	link := n.(*ast.Link)
	_, _ = w.WriteString(`<a href="`)
	dest := util.URLEscape(link.Destination, true)
	_, _ = w.Write(util.EscapeHTML(dest))
	_, _ = w.WriteString(`"`)
	if len(link.Title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML(link.Title))
		_, _ = w.WriteString(`"`)
	}
	if opensInNewTab(link.Destination) {
		_, _ = w.WriteString(` target="_blank" rel="noopener noreferrer"`)
	}
	if link.Attributes() != nil {
		gmhtml.RenderAttributes(w, link, gmhtml.LinkAttributeFilter)
	}
	_ = w.WriteByte('>')
	return ast.WalkContinue, nil
}

func opensInNewTab(destination []byte) bool {
	lower := strings.ToLower(string(destination))
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "file:") ||
		strings.HasPrefix(lower, "//")
}

func (r *safeLinkRenderer) renderImage(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	image := n.(*ast.Image)
	_, _ = w.WriteString(`<img src="`)
	dest := util.URLEscape(image.Destination, true)
	_, _ = w.Write(util.EscapeHTML(dest))
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(util.EscapeHTML([]byte(nodeText(image, source))))
	_ = w.WriteByte('"')
	if len(image.Title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML(image.Title))
		_, _ = w.WriteString(`"`)
	}
	if image.Attributes() != nil {
		gmhtml.RenderAttributes(w, image, gmhtml.ImageAttributeFilter)
	}
	_ = w.WriteByte('>')
	return ast.WalkSkipChildren, nil
}
