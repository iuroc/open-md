package render

import (
	"net/url"
	"path/filepath"
	"strings"
)

type URLKind int

const (
	LinkURL URLKind = iota
	ImageURL
)

func RewriteURL(raw, baseDir string, kind URLKind, knownDocs map[string]string) (string, error) {
	if raw == "" {
		return "", nil
	}

	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "vbscript:") {
		return "", nil
	}
	if strings.HasPrefix(raw, "#") {
		return raw, nil
	}

	if filepath.IsAbs(raw) {
		abs := filepath.Clean(raw)
		if kind == LinkURL {
			if id, ok := lookupKnown(abs, knownDocs); ok {
				return docHash(id, ""), nil
			}
		}
		return fileURL(abs, "", "")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw, err
	}
	if parsed.Scheme != "" {
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https", "file":
			return raw, nil
		case "mailto", "tel":
			if kind == LinkURL {
				return raw, nil
			}
			return "", nil
		case "data":
			if kind == ImageURL && strings.HasPrefix(lower, "data:image/") {
				return raw, nil
			}
			return "", nil
		default:
			return "", nil
		}
	}
	if parsed.Host != "" || parsed.Path == "" {
		return raw, nil
	}

	localPath := filepath.FromSlash(parsed.Path)
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(baseDir, localPath)
	}
	abs, err := filepath.Abs(localPath)
	if err != nil {
		return raw, err
	}
	if kind == LinkURL {
		if id, ok := lookupKnown(abs, knownDocs); ok {
			return docHash(id, parsed.Fragment), nil
		}
	}
	return fileURL(abs, parsed.RawQuery, parsed.Fragment)
}

func lookupKnown(abs string, knownDocs map[string]string) (string, bool) {
	if knownDocs == nil {
		return "", false
	}
	id, ok := knownDocs[strings.ToLower(filepath.Clean(abs))]
	return id, ok
}

func docHash(id, fragment string) string {
	hash := "#doc=" + url.QueryEscape(id)
	if fragment != "" {
		hash += "&anchor=" + url.QueryEscape(fragment)
	}
	return hash
}

func fileURL(abs, rawQuery, fragment string) (string, error) {
	abs = filepath.Clean(abs)
	if strings.HasPrefix(abs, `\\`) {
		rest := strings.TrimPrefix(abs, `\\`)
		parts := strings.SplitN(rest, `\`, 2)
		path := "/"
		if len(parts) == 2 {
			path += filepath.ToSlash(parts[1])
		}
		u := &url.URL{
			Scheme:   "file",
			Host:     parts[0],
			Path:     path,
			RawQuery: rawQuery,
			Fragment: fragment,
		}
		return u.String(), nil
	}
	u := &url.URL{
		Scheme:   "file",
		Path:     "/" + filepath.ToSlash(abs),
		RawQuery: rawQuery,
		Fragment: fragment,
	}
	return u.String(), nil
}
