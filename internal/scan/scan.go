package scan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Limits struct {
	MaxDocs       int
	MaxTotalBytes int64
	MaxFileBytes  int64
	MaxDepth      int
	MaxEntries    int
}

var DefaultLimits = Limits{
	MaxDocs:       500,
	MaxTotalBytes: 25 << 20,
	MaxFileBytes:  32 << 20,
	MaxDepth:      12,
	MaxEntries:    20000,
}

// QuickLimits applies the "quick mode" selection used when the user declines a
// full large-directory load.
var QuickLimits = Limits{
	MaxDocs:       3000,
	MaxTotalBytes: 8 << 20,
	MaxFileBytes:  32 << 20,
	MaxDepth:      12,
	MaxEntries:    100000,
}

// UnlimitedLimits removes business caps while still skipping symlinks,
// reparse points, unreadable directories, and known heavy directories.
var UnlimitedLimits = Limits{
	MaxDocs:       -1,
	MaxTotalBytes: -1,
	MaxFileBytes:  -1,
	MaxDepth:      -1,
	MaxEntries:    -1,
}

type Document struct {
	Path    string
	Name    string
	RelPath string
	Size    int64
	ModTime int64
}

type SkippedDocument struct {
	Path    string
	Name    string
	RelPath string
	Size    int64
}

type Result struct {
	Documents []Document
	Skipped   []SkippedDocument
	Warnings  []string
	Truncated bool
}

var (
	ErrNoMarkdown         = errors.New("目录中没有 Markdown 文件")
	ErrNoLoadableMarkdown = errors.New("目录中的 Markdown 文件均超过大小限制，无法加载")
)

var ignoredDirectories = map[string]bool{
	".git":                      true,
	".svn":                      true,
	".hg":                       true,
	"node_modules":              true,
	"vendor":                    true,
	".obsidian":                 true,
	".trash":                    true,
	"$recycle.bin":              true,
	"system volume information": true,
}

type candidate struct {
	path    string
	name    string
	relPath string
	size    int64
	modTime int64
	depth   int
}

func Directory(dir string, limits Limits) (*Result, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("无法解析目录路径: %w", err)
	}
	abs = filepath.Clean(abs)
	limits = withDefaults(limits)
	perFileBudget := limits.MaxFileBytes
	if perFileBudget > 0 && limits.MaxTotalBytes > 0 && perFileBudget > limits.MaxTotalBytes {
		perFileBudget = limits.MaxTotalBytes
	}

	result := &Result{}
	candidates := make([]candidate, 0)
	visited := 0
	stopped := false
	depthWarned := false

	var walk func(current string, depth int) error
	walk = func(current string, depth int) error {
		entries, err := os.ReadDir(current)
		if err != nil {
			if depth == 0 {
				return fmt.Errorf("无法读取目录: %w", err)
			}
			result.Warnings = append(result.Warnings, "无法读取目录："+relPath(abs, current))
			return nil
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return naturalLess(entries[i].Name(), entries[j].Name())
		})
		for _, entry := range entries {
			if stopped {
				return nil
			}
			visited++
			if limits.MaxEntries >= 0 && visited > limits.MaxEntries {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"目录项过多，扫描范围已截断（上限 %d 项）", limits.MaxEntries,
				))
				stopped = true
				return nil
			}
			name := entry.Name()
			full := filepath.Join(current, name)
			if isReparsePoint(full) {
				continue
			}
			if entry.IsDir() {
				if ignoredDirectories[strings.ToLower(name)] {
					continue
				}
				if limits.MaxDepth >= 0 && depth+1 > limits.MaxDepth {
					if !depthWarned {
						result.Warnings = append(result.Warnings, fmt.Sprintf(
							"部分子目录超过最大扫描深度（%d 层），已跳过", limits.MaxDepth,
						))
						depthWarned = true
					}
					continue
				}
				if err := walk(full, depth+1); err != nil {
					return err
				}
				continue
			}
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".md" && ext != ".markdown" {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				result.Warnings = append(result.Warnings, "无法读取文件信息："+relPath(abs, full))
				continue
			}
			rel := relPath(abs, full)
			if perFileBudget > 0 && info.Size() > perFileBudget {
				result.Skipped = append(result.Skipped, SkippedDocument{
					Path:    full,
					Name:    name,
					RelPath: rel,
					Size:    info.Size(),
				})
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"文件过大，已跳过：%s（%s）", rel, formatSize(info.Size()),
				))
				continue
			}
			candidates = append(candidates, candidate{
				path:    full,
				name:    name,
				relPath: rel,
				size:    info.Size(),
				modTime: info.ModTime().UnixNano(),
				depth:   depth,
			})
		}
		return nil
	}

	if err := walk(abs, 0); err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		if len(result.Skipped) > 0 {
			return nil, ErrNoLoadableMarkdown
		}
		return nil, ErrNoMarkdown
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].depth != candidates[j].depth {
			return candidates[i].depth < candidates[j].depth
		}
		return naturalPathLess(candidates[i].relPath, candidates[j].relPath)
	})

	var total int64
	for _, c := range candidates {
		if limits.MaxDocs >= 0 && len(result.Documents) >= limits.MaxDocs {
			result.Truncated = true
			break
		}
		if limits.MaxTotalBytes >= 0 && total+c.size > limits.MaxTotalBytes {
			result.Truncated = true
			break
		}
		total += c.size
		result.Documents = append(result.Documents, Document{
			Path:    c.path,
			Name:    c.name,
			RelPath: c.relPath,
			Size:    c.size,
			ModTime: c.modTime,
		})
	}
	return result, nil
}

func withDefaults(limits Limits) Limits {
	if limits.MaxDocs == 0 {
		limits.MaxDocs = DefaultLimits.MaxDocs
	}
	if limits.MaxTotalBytes == 0 {
		limits.MaxTotalBytes = DefaultLimits.MaxTotalBytes
	}
	if limits.MaxFileBytes == 0 {
		limits.MaxFileBytes = DefaultLimits.MaxFileBytes
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = DefaultLimits.MaxDepth
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = DefaultLimits.MaxEntries
	}
	return limits
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func naturalPathLess(a, b string) bool {
	as := strings.Split(a, "/")
	bs := strings.Split(b, "/")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] == bs[i] {
			continue
		}
		return naturalLess(as[i], bs[i])
	}
	return len(as) < len(bs)
}

// NaturalLess reports whether a sorts before b using the project's natural order.
func NaturalLess(a, b string) bool {
	return naturalLess(a, b)
}

func naturalLess(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if isDigit(a[i]) && isDigit(b[j]) {
			iStart, jStart := i, j
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			aDigits := strings.TrimLeft(a[iStart:i], "0")
			bDigits := strings.TrimLeft(b[jStart:j], "0")
			if len(aDigits) != len(bDigits) {
				return len(aDigits) < len(bDigits)
			}
			if aDigits != bDigits {
				return aDigits < bDigits
			}
			continue
		}
		if a[i] != b[j] {
			return a[i] < b[j]
		}
		i++
		j++
	}
	return len(a) < len(b)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func formatSize(size int64) string {
	if size >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	}
	if size >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(size)/(1<<10))
	}
	return fmt.Sprintf("%d B", size)
}
