package manual

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

var chapterFiles = []string{
	"01-what-is-markdown.md",
	"02-headings-paragraphs.md",
	"03-emphasis-inline-code.md",
	"04-lists.md",
	"05-links-images.md",
	"06-blockquotes.md",
	"07-code-blocks.md",
	"08-tables.md",
	"09-rules-and-html.md",
	"10-organize-documents.md",
	"11-recipes.md",
	"12-cheatsheet.md",
	"13-faq.md",
}

var inlineCode = regexp.MustCompile("`[^`\n]*`")

// prose removes fenced code blocks and inline code so syntax examples do not
// count as real headings, links, or further-reading citations.
func prose(text string) string {
	lines := strings.Split(text, "\n")
	var b strings.Builder
	inFence := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			b.WriteString("\n")
			continue
		}
		if inFence {
			b.WriteString("\n")
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return inlineCode.ReplaceAllString(b.String(), "")
}

func TestManualContentStructure(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		text := string(data)
		if len(regexp.MustCompile(`(?m)^# `).FindAllString(prose(text), -1)) != 1 {
			t.Fatalf("%s must have exactly one H1", name)
		}
		if name == "index.md" {
			continue
		}
		for _, heading := range []string{
			"## 你会学到",
			"## 基本写法",
			"## 在 OpenMD 里怎么看",
			"## 常见错误",
			"## 延伸阅读",
		} {
			if !strings.Contains(prose(text), heading) {
				t.Fatalf("%s missing %q", name, heading)
			}
		}
		if !strings.Contains(text, "```") {
			t.Fatalf("%s must contain a fenced example", name)
		}
		if !strings.Contains(prose(text), "https://") {
			t.Fatalf("%s needs a further-reading link", name)
		}
	}
}

func TestManualIndexLinksEveryChapter(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(fsys, "index.md")
	if err != nil {
		t.Fatal(err)
	}
	text := prose(string(data))
	for _, name := range chapterFiles {
		if !strings.Contains(text, "]("+name+")") {
			t.Fatalf("index.md does not link to %s", name)
		}
	}
}

var markdownLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

func TestManualRelativeLinksResolve(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range markdownLink.FindAllStringSubmatch(prose(string(data)), -1) {
			target := match[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			if idx := strings.Index(target, "#"); idx >= 0 {
				target = target[:idx]
			}
			if !strings.HasSuffix(target, ".md") {
				continue
			}
			resolved := path.Clean(path.Join(path.Dir(name), target))
			if _, err := fs.Stat(fsys, resolved); err != nil {
				t.Fatalf("%s links to missing %s", name, resolved)
			}
		}
	}
}

var imageSyntax = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`)

func TestManualProseImagesExist(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range imageSyntax.FindAllStringSubmatch(prose(string(data)), -1) {
			target := match[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "data:") || strings.HasPrefix(target, "#") {
				continue
			}
			resolved := path.Clean(path.Join(path.Dir(name), target))
			if _, err := fs.Stat(fsys, resolved); err != nil {
				t.Fatalf("%s renders a missing image: %s", name, resolved)
			}
		}
	}
}

func TestManualHasNoRemoteImages(t *testing.T) {
	fsys, err := contentFS()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range append([]string{"index.md"}, chapterFiles...) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, "<img") || regexp.MustCompile(`!\[[^\]]*\]\(https?://`).MatchString(text) {
			t.Fatalf("%s must not reference remote images", name)
		}
	}
}
