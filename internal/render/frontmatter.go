package render

import "bytes"

const (
	maxFrontMatterBytes = 64 << 10
	maxFrontMatterLines = 200
)

// stripFrontMatter removes a leading YAML (---) or TOML (+++) metadata block.
// It deliberately does not parse the metadata; it only recognizes a bounded
// block that has a plausible key/value shape. This keeps ordinary horizontal
// rules and malformed input as normal Markdown.
func stripFrontMatter(source []byte) []byte {
	if len(source) == 0 {
		return source
	}

	firstLineEnd := bytes.IndexByte(source, '\n')
	if firstLineEnd < 0 {
		return source
	}
	delimiter := string(bytes.TrimRight(source[:firstLineEnd], "\r"))
	if delimiter != "---" && delimiter != "+++" {
		return source
	}

	contentStart := firstLineEnd + 1
	offset := contentStart
	scanned := 0
	lines := 0
	for offset <= len(source) && scanned <= maxFrontMatterBytes && lines < maxFrontMatterLines {
		lineEnd := bytes.IndexByte(source[offset:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += offset
		}
		line := string(bytes.TrimRight(source[offset:lineEnd], "\r"))
		if line == delimiter {
			if !hasFrontMatterField(source[contentStart:offset], delimiter) {
				return source
			}
			if lineEnd < len(source) {
				lineEnd++
			}
			return source[lineEnd:]
		}
		scanned += lineEnd - offset + 1
		lines++
		if lineEnd >= len(source) {
			break
		}
		offset = lineEnd + 1
	}
	return source
}

func hasFrontMatterField(block []byte, delimiter string) bool {
	for _, rawLine := range bytes.Split(block, []byte("\n")) {
		line := string(bytes.TrimSpace(rawLine))
		if line == "" || bytes.HasPrefix([]byte(line), []byte("#")) {
			continue
		}
		if delimiter == "+++" {
			if len(line) >= 2 && line[0] == '[' && line[len(line)-1] == ']' {
				return true
			}
			if equal := bytes.IndexByte([]byte(line), '='); equal > 0 {
				return true
			}
			continue
		}
		if colon := bytes.IndexByte([]byte(line), ':'); colon > 0 {
			return true
		}
	}
	return false
}
