package render

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type safeLinkRenderer struct{}

func (r *safeLinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindImage, r.renderImage)
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
	if link.Attributes() != nil {
		gmhtml.RenderAttributes(w, link, gmhtml.LinkAttributeFilter)
	}
	_ = w.WriteByte('>')
	return ast.WalkContinue, nil
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
