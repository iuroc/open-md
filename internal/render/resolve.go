package render

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type localTarget struct {
	Path     string
	RawQuery string
	Fragment string
	Display  string
}

// resolveLocalDestination resolves a Markdown destination against the source
// directory. Existing literal filenames win over URL semantics so names with
// %, # or other URL-significant characters keep working. URL decoding is used
// only as a fallback. Network paths are never stat'ed here.
func resolveLocalDestination(raw, baseDir string) (localTarget, bool) {
	if raw == "" {
		return localTarget{}, false
	}
	if filepath.IsAbs(raw) {
		path := filepath.Clean(raw)
		return localTarget{Path: path, Display: path}, true
	}

	literal := filepath.FromSlash(raw)
	if !filepath.IsAbs(literal) {
		literal = filepath.Join(baseDir, literal)
	}
	if abs, err := filepath.Abs(literal); err == nil {
		literal = abs
	}
	if !isNetworkPath(literal) {
		if info, err := os.Stat(literal); err == nil && !info.IsDir() {
			return localTarget{Path: literal, Display: displayPath(raw)}, true
		}
	}

	// A literal filename may contain # on Windows, so only split after the
	// whole filename has been ruled out. This also rescues names such as
	// "100%.md#section", where url.Parse would reject the % escape.
	if cut := strings.IndexAny(raw, "?#"); cut > 0 {
		if target, ok := resolveLocalDestination(raw[:cut], baseDir); ok {
			rest := raw[cut:]
			if strings.HasPrefix(rest, "?") {
				query := rest[1:]
				if hash := strings.IndexByte(query, '#'); hash >= 0 {
					target.RawQuery = query[:hash]
					target.Fragment = decodeFragment(query[hash+1:])
				} else {
					target.RawQuery = query
				}
			} else {
				target.Fragment = decodeFragment(rest[1:])
			}
			return target, true
		}
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return localTarget{Path: literal, Display: displayPath(raw)}, true
	}
	switch strings.ToLower(parsed.Scheme) {
	case "":
		if parsed.Host != "" || parsed.Path == "" || !utf8.ValidString(parsed.Path) {
			return localTarget{}, false
		}
		local := filepath.FromSlash(parsed.Path)
		if !filepath.IsAbs(local) {
			local = filepath.Join(baseDir, local)
		}
		abs, absErr := filepath.Abs(local)
		if absErr != nil {
			return localTarget{}, false
		}
		return localTarget{
			Path:     abs,
			RawQuery: parsed.RawQuery,
			Fragment: parsed.Fragment,
			Display:  displayPath(parsed.Path),
		}, true
	case "file":
		if parsed.Host != "" || parsed.Path == "" {
			return localTarget{}, false
		}
		local := filepath.FromSlash(parsed.Path)
		if len(local) >= 3 && local[0] == '\\' && local[2] == ':' {
			local = local[1:]
		}
		abs, absErr := filepath.Abs(local)
		if absErr != nil {
			return localTarget{}, false
		}
		return localTarget{
			Path:     abs,
			RawQuery: parsed.RawQuery,
			Fragment: parsed.Fragment,
			Display:  abs,
		}, true
	default:
		return localTarget{}, false
	}
}

func displayPath(raw string) string {
	display := filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw)))
	if display == "." {
		return ""
	}
	return display
}
