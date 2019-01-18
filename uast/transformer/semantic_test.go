package transformer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComment(t *testing.T) {
	var cases = []struct {
		name string
		text string
		exp  commentElems
	}{
		{
			name: "one line",
			text: "// some text",
			exp: commentElems{
				Tokens: [2]string{"//", ""},
				Pref:   " ",
				Text:   "some text",
			},
		},
		{
			name: "new line",
			text: "// some text\n",
			exp: commentElems{
				Tokens: [2]string{"//", ""},
				Pref:   " ", Suff: "\n",
				Text: "some text",
			},
		},
		{
			name: "multi-line single",
			text: "/* some text */",
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   " ", Suff: " ",
				Text: "some text",
			},
		},
		{
			name: "multi-line new line",
			text: `/*
	some text
*/`,
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   "\n\t", Suff: "\n",
				Text: "some text",
			},
		},
		{
			name: "multi-line",
			text: `/*
	some text
	line two
*/`,
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   "\n\t", Tab: "\t", Suff: "\n",
				Text: "some text\nline two",
			},
		},
		{
			name: "stylistic",
			text: `/*
 * some text
 * line two
 * line three
*/`,
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   "\n * ", Tab: " * ", Suff: "\n",
				Text: "some text\nline two\nline three",
			},
		},
		{
			name: "multiple single line",
			text: `// some text
// line two`,
			exp: commentElems{
				Tokens: [2]string{"//", ""},
				Pref:   " ", Tab: "// ", Suff: "",
				Text: "some text\nline two",
			},
		},
		{
			name: "stylistic inconsistent",
			text: `/*
 * some text
 *   line two
 * line three
*/`,
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   "\n * ", Tab: " * ", Suff: "\n",
				Text: "some text\n  line two\nline three",
			},
		},
		{
			name: "inconsistent",
			text: `/*
 * some text
   line two
 * line three
*/`,
			exp: commentElems{
				Tokens: [2]string{"/*", "*/"},
				Pref:   "\n * ", Tab: " ", Suff: "\n",
				Text: "some text\n  line two\n* line three",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := commentElems{Tokens: c.exp.Tokens}
			if !v.Split(c.text) {
				t.Error("split failed")
			}
			require.Equal(t, c.exp, v)
			require.Equal(t, c.text, v.Join())
		})
	}
}
