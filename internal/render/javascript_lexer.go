package render

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// The bundled Chroma JavaScript lexer predates private class fields (ES2022),
// so a leading '#' in "#items" falls through to the error token and picks up
// the theme's error background. Patch the root state so private names are
// highlighted as ordinary identifiers.
func init() {
	base, ok := lexers.Get("javascript").(*chroma.RegexLexer)
	if !ok {
		return
	}
	rules, err := base.Rules()
	if err != nil {
		return
	}
	patched := make(chroma.Rules, len(rules))
	for state, stateRules := range rules {
		patched[state] = append([]chroma.Rule(nil), stateRules...)
	}
	root := patched["root"]
	patched["root"] = append([]chroma.Rule{{
		Pattern: `#[_\p{L}$][_\p{L}\p{N}$]*`,
		Type:    chroma.NameOther,
	}}, root...)

	config := *base.Config()
	lexers.Register(chroma.MustNewLexer(&config, func() chroma.Rules {
		return patched
	}))
}
