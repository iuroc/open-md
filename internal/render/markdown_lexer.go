package render

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// The custom lexer keeps Chroma's Markdown behavior, adds a few common GFM
// forms that Chroma does not recognize yet, and handles both YAML and TOML
// front matter. It is registered under the name "markdown" so goldmark's
// ```markdown fences and the source view use the same rules.
func init() {
	base := chroma.MustNewLexer(&chroma.Config{
		Name:      "markdown",
		Aliases:   []string{"md", "mkd"},
		Filenames: []string{"*.md", "*.mkd", "*.markdown"},
		MimeTypes: []string{"text/x-markdown"},
	}, openmdMarkdownRules)
	lexers.Register(&markdownFrontMatterLexer{Lexer: base})
	registerLessAlias()
}

func registerLessAlias() {
	css := lexers.Get("css")
	if css == nil {
		return
	}
	config := css.Config()
	for _, alias := range config.Aliases {
		if alias == "less" {
			return
		}
	}
	config.Aliases = append(config.Aliases, "less")
	lexers.Register(css)
}

func openmdMarkdownRules() chroma.Rules {
	return chroma.Rules{
		"root": {
			{`<!--[\w\W]*?-->`, chroma.CommentMultiline, nil},

			// Fenced code first: it consumes its whole body so the rules
			// below cannot reinterpret code as Markdown syntax.
			{"^(```)(\\w+)(\\n)([\\w\\W]*?)(^```$)", chroma.UsingByGroup(2, 4, chroma.String, chroma.String, chroma.String, chroma.Text, chroma.String), nil},
			{"^(```\\n)([\\w\\W]*?)(^```$)", chroma.ByGroups(chroma.String, chroma.Text, chroma.String), nil},

			{`^(#[^#].+\n)`, chroma.GenericHeading, nil},
			{`^(#{2,6}.+\n)`, chroma.GenericSubheading, nil},

			// GFM tables that start with a pipe. Four or more leading spaces
			// are left alone so indented code is not mistaken for a table.
			{`^( {0,3})(\|)`, chroma.ByGroups(chroma.Text, chroma.Operator), chroma.Push("table")},

			{`^([ \t]*\[\^[^\]\n]+\]:)(.*\n?)`, chroma.ByGroups(chroma.NameEntity, chroma.UsingSelf("inline")), nil},
			{`^([ \t]*\[)([^\]\n]+)(\]:)([ \t]*)(\S+)(.*)$`, chroma.ByGroups(chroma.Text, chroma.NameTag, chroma.Punctuation, chroma.Text, chroma.NameAttribute, chroma.Text), nil},

			{`^(\s*)([*-] )(\[[ xX]\])( .+\n)`, chroma.ByGroups(chroma.Text, chroma.Keyword, chroma.Keyword, chroma.UsingSelf("inline")), nil},
			{`^(\s*)([*-])(\s)(.+\n)`, chroma.ByGroups(chroma.Text, chroma.Keyword, chroma.Text, chroma.UsingSelf("inline")), nil},
			{`^(\s*)([0-9]+\.)( .+\n)`, chroma.ByGroups(chroma.Text, chroma.Keyword, chroma.UsingSelf("inline")), nil},
			{`^(\s*>\s)(.+\n)`, chroma.ByGroups(chroma.Keyword, chroma.GenericEmph), nil},

			// Setext headings are checked after line prefixes so a list item
			// followed by --- is not mistaken for a heading.
			{`^(.+)(\n)(=+[ \t]*\n)`, chroma.ByGroups(chroma.GenericHeading, chroma.Text, chroma.GenericHeading), nil},
			{`^(.+)(\n)(-+[ \t]*\n)`, chroma.ByGroups(chroma.GenericSubheading, chroma.Text, chroma.GenericSubheading), nil},

			chroma.Include("inline"),
		},
		"table": {
			{`\|`, chroma.Operator, nil},
			{`[ \t]*:?-{3,}:?[ \t]*`, chroma.Keyword, nil},
			{`[^|\n]+`, chroma.UsingSelf("inline"), nil},
			{`\n`, chroma.Text, chroma.Pop(1)},
		},
		"inline": {
			{`<!--[\w\W]*?-->`, chroma.CommentMultiline, nil},
			{`\\.`, chroma.Text, nil},

			// Autolinks must be checked before generic HTML tags.
			{`(<)((?:https?|ftp|mailto):[^>\n]+)(>)`, chroma.ByGroups(chroma.Punctuation, chroma.NameAttribute, chroma.Punctuation), nil},
			{`(<)([^<>\s@]+@[^<>\s@]+\.[^<>\s@]+)(>)`, chroma.ByGroups(chroma.Punctuation, chroma.NameAttribute, chroma.Punctuation), nil},
			{`(</?)([A-Za-z][\w:-]*)([^>\n]*)(/?>)`, chroma.ByGroups(chroma.Punctuation, chroma.NameTag, chroma.NameAttribute, chroma.Punctuation), nil},

			{`(\s)(\*|_)((?:(?!\2).)*)(\2)((?=\W|\n))`, chroma.ByGroups(chroma.Text, chroma.GenericEmph, chroma.GenericEmph, chroma.GenericEmph, chroma.Text), nil},
			{`(\s)((\*\*|__).*?)\3((?=\W|\n))`, chroma.ByGroups(chroma.Text, chroma.GenericStrong, chroma.GenericStrong, chroma.Text), nil},
			{`(\s)(~~[^~]+~~)((?=\W|\n))`, chroma.ByGroups(chroma.Text, chroma.GenericDeleted, chroma.Text), nil},
			{"`[^`]+`", chroma.LiteralStringBacktick, nil},

			{`\[\^[^\]\n]+\]`, chroma.NameEntity, nil},
			{`(!?\[)([^\]\n]+)(\])(\[)([^\]\n]*)(\])`, chroma.ByGroups(chroma.Text, chroma.NameTag, chroma.Text, chroma.Text, chroma.NameAttribute, chroma.Text), nil},
			{`(!?\[)([^]]+)(\])(\()([^)]+)(\))`, chroma.ByGroups(chroma.Text, chroma.NameTag, chroma.Text, chroma.Text, chroma.NameAttribute, chroma.Text), nil},
			{`[@#][\w/:]+`, chroma.NameEntity, nil},
			{`.|\n`, chroma.Text, nil},
		},
	}
}

// markdownFrontMatterLexer delegates to the Markdown rules after highlighting
// a leading YAML or TOML front matter block.
type markdownFrontMatterLexer struct {
	chroma.Lexer
}

func (l *markdownFrontMatterLexer) Tokenise(options *chroma.TokeniseOptions, text string) (chroma.Iterator, error) {
	fm, ok := splitFrontMatterBlock(text)
	if !ok {
		return l.Lexer.Tokenise(options, text)
	}
	rest, err := l.Lexer.Tokenise(options, fm.rest)
	if err != nil {
		return nil, err
	}
	var front chroma.Iterator
	switch fm.lang {
	case "TOML":
		tomlLexer := lexers.Get("TOML")
		if tomlLexer == nil {
			return l.Lexer.Tokenise(options, text)
		}
		body, err := tomlLexer.Tokenise(options, fm.body)
		if err != nil {
			return nil, err
		}
		front = chroma.Concaterator(
			chroma.Literator(chroma.Token{Type: chroma.Operator, Value: fm.open}),
			body,
			chroma.Literator(chroma.Token{Type: chroma.Operator, Value: fm.close}),
		)
	default:
		yamlLexer := lexers.Get("YAML")
		if yamlLexer == nil {
			return l.Lexer.Tokenise(options, text)
		}
		front, err = yamlLexer.Tokenise(options, fm.open+fm.body+fm.close)
		if err != nil {
			return nil, err
		}
	}
	return chroma.Concaterator(front, rest), nil
}

type frontMatterBlock struct {
	open  string
	body  string
	close string
	rest  string
	lang  string
}

func splitFrontMatterBlock(text string) (frontMatterBlock, bool) {
	for _, spec := range []struct {
		delim string
		lang  string
	}{
		{"---", "YAML"},
		{"+++", "TOML"},
	} {
		open := spec.delim + "\n"
		if strings.HasPrefix(text, spec.delim+"\r\n") {
			open = spec.delim + "\r\n"
		}
		if !strings.HasPrefix(text, open) {
			continue
		}
		pos := len(open)
		for pos <= len(text) {
			next := strings.IndexByte(text[pos:], '\n')
			if next < 0 {
				if strings.TrimSuffix(text[pos:], "\r") == spec.delim {
					return frontMatterBlock{
						open:  text[:len(open)],
						body:  text[len(open):pos],
						close: text[pos:],
						rest:  "",
						lang:  spec.lang,
					}, true
				}
				break
			}
			lineEnd := pos + next
			if strings.TrimSuffix(text[pos:lineEnd], "\r") == spec.delim {
				closeEnd := lineEnd + 1
				return frontMatterBlock{
					open:  text[:len(open)],
					body:  text[len(open):pos],
					close: text[pos:closeEnd],
					rest:  text[closeEnd:],
					lang:  spec.lang,
				}, true
			}
			pos = lineEnd + 1
		}
	}
	return frontMatterBlock{}, false
}
