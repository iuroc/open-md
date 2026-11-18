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

func RewriteURL(raw, baseDir string, kind URLKind, knownDocs map[string]string, currentDocID string) (string, error) {
	if raw == "" {
		return "", nil
	}

	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "vbscript:") {
		return "", nil
	}
	if strings.HasPrefix(raw, "#") {
		if currentDocID != "" {
			return docHash(currentDocID, decodeFragment(strings.TrimPrefix(raw, "#"))), nil
		}
		return raw, nil
	}

	if filepath.IsAbs(raw) {
		return rewriteLocalTarget(localTarget{Path: filepath.Clean(raw)}, kind, knownDocs)
	}

	if parsed, err := url.Parse(raw); err == nil {
		switch strings.ToLower(parsed.Scheme) {
		case "":
			if parsed.Host != "" || parsed.Path == "" {
				return raw, nil
			}
		case "http", "https":
			return raw, nil
		case "file":
			// UNC and other host-bearing file URLs stay as-is: never stat them.
			if parsed.Host != "" {
				return raw, nil
			}
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

	target, ok := resolveLocalDestination(raw, baseDir)
	if !ok {
		return "", nil
	}
	return rewriteLocalTarget(target, kind, knownDocs)
}

func rewriteLocalTarget(target localTarget, kind URLKind, knownDocs map[string]string) (string, error) {
	if kind == LinkURL {
		if id, ok := lookupKnown(target.Path, knownDocs); ok {
			return docHash(id, target.Fragment), nil
		}
	}
	return fileURL(target.Path, target.RawQuery, target.Fragment)
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

func decodeFragment(fragment string) string {
	decoded, err := url.PathUnescape(fragment)
	if err != nil {
		return fragment
	}
	return decoded
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
