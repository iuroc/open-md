package render

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"html/template"

	"openmd/internal/buildinfo"
)

//go:embed web/docpage.html
var docPageTemplateSource string

var docPageTemplate = template.Must(template.New("docpage").Parse(docPageTemplateSource))

type DocPageData struct {
	Title            string
	Mode             string
	RootKey          string
	HasFilesPanel    bool
	Doc              Document
	VersionLabel     string
	BuildTime        string
	AuthorName       string
	AuthorURL        string
	ReleaseAvailable bool
	ReleaseURL       string
	Favicon          template.URL
}

func prepareDocPageData(data DocPageData) DocPageData {
	if data.Title == "" {
		data.Title = data.Doc.Title
	}
	if data.VersionLabel == "" {
		data.VersionLabel = buildinfo.VersionLabel()
	}
	if data.BuildTime == "" {
		data.BuildTime = buildinfo.BuildTimeDisplay()
	}
	if data.AuthorName == "" {
		data.AuthorName = buildinfo.AuthorName
	}
	if data.AuthorURL == "" {
		data.AuthorURL = buildinfo.AuthorURL
	}
	if data.Favicon == "" {
		data.Favicon = template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(faviconPNG))
	}
	data.ReleaseAvailable = buildinfo.IsReleaseBuild()
	data.ReleaseURL = buildinfo.ReleaseURL
	return data
}

func RenderDocPage(data DocPageData) ([]byte, error) {
	data = prepareDocPageData(data)

	var buf bytes.Buffer
	if err := docPageTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func RenderDocSection(data DocPageData) ([]byte, error) {
	data = prepareDocPageData(data)

	var buf bytes.Buffer
	if err := docPageTemplate.ExecuteTemplate(&buf, "docsection", data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
