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
}

var DefaultLimits = Limits{MaxDocs: 500, MaxTotalBytes: 25 << 20, MaxFileBytes: 32 << 20}

type Document struct {
	Path string
	Name string
	Size int64
}

type SkippedDocument struct {
	Path string
	Name string
	Size int64
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

func Directory(dir string, limits Limits) (*Result, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("无法解析目录路径: %w", err)
	}
	if limits.MaxDocs <= 0 {
		limits.MaxDocs = DefaultLimits.MaxDocs
	}
	if limits.MaxTotalBytes <= 0 {
		limits.MaxTotalBytes = DefaultLimits.MaxTotalBytes
	}
	if limits.MaxFileBytes <= 0 {
		limits.MaxFileBytes = DefaultLimits.MaxFileBytes
	}
	perFileBudget := limits.MaxFileBytes
	if perFileBudget > limits.MaxTotalBytes {
		perFileBudget = limits.MaxTotalBytes
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("无法读取目录: %w", err)
	}

	result := &Result{}
	var total int64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".md" && ext != ".markdown" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			result.Warnings = append(result.Warnings, "无法读取文件信息："+entry.Name())
			continue
		}
		if len(result.Documents) >= limits.MaxDocs {
			result.Truncated = true
			break
		}
		if info.Size() > perFileBudget {
			result.Skipped = append(result.Skipped, SkippedDocument{
				Path: filepath.Join(abs, entry.Name()),
				Name: entry.Name(),
				Size: info.Size(),
			})
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"文件过大，已跳过：%s（%s）", entry.Name(), formatSize(info.Size()),
			))
			continue
		}
		if total+info.Size() > limits.MaxTotalBytes {
			result.Truncated = true
			break
		}
		total += info.Size()
		result.Documents = append(result.Documents, Document{
			Path: filepath.Join(abs, entry.Name()),
			Name: entry.Name(),
			Size: info.Size(),
		})
	}

	if len(result.Documents) == 0 {
		if len(result.Skipped) > 0 {
			return nil, ErrNoLoadableMarkdown
		}
		return nil, ErrNoMarkdown
	}
	sort.SliceStable(result.Documents, func(i, j int) bool {
		return naturalLess(result.Documents[i].Name, result.Documents[j].Name)
	})
	return result, nil
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
