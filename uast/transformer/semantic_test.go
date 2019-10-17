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
				StartToken: "//", EndToken: "",
				Prefix: " ",
				Text:   "some text",
			},
		},
		{
			name: "space",
			text: "// ",
			exp: commentElems{
				StartToken: "//", EndToken: "",
				Prefix: " ",
				Text:   "",
			},
		},
		{
			name: "utf8 chars",
			text: "// “utf8 ½Å commentÅ”",
			exp: commentElems{
				StartToken: "//", EndToken: "",
				Prefix: " ", Suffix: "",
				Text:   "“utf8 ½Å commentÅ”",
			},
		},
		{
			name: "utf8 singleton",
			text: "/*\t\t\u00a0 ½ */",
			exp: commentElems{
				StartToken: "/*", EndToken: "",
				Prefix:"\t\t\u00a0 ", Suffix:" */",
				Text:   "½",
			},
		},
		{
			name: "new line",
			text: "// some text\n",
			exp: commentElems{
				StartToken: "//", EndToken: "",
				Prefix: " ", Suffix: "\n",
				Text: "some text",
			},
		},
		{
			name: "multi-line single",
			text: "/* some text */",
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: " ", Suffix: " ",
				Text: "some text",
			},
		},
		{
			name: "multi-line new line",
			text:
`/*
	some text
*/`,
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: "\n\t", Suffix: "\n",
				Text: "some text",
			},
		},
		{
			name: "multi-line",
			text:
`/*
	some text
	line two
*/`,
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: "\n\t", Indent: "\t", Suffix: "\n",
				Text: "some text\nline two",
			},
		},
		{
			name: "stylistic",
			text:
`/*
 * some text
 * line two
 * line three
*/`,
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: "\n * ", Indent: " * ", Suffix: "\n",
				Text: "some text\nline two\nline three",
			},
		},
		{
			name: "multiple single line",
			text:
`// some text
// line two`,
			exp: commentElems{
				StartToken: "//", EndToken: "",
				Prefix: " ", Indent: "// ", Suffix: "",
				Text: "some text\nline two",
			},
		},
		{
			name: "stylistic inconsistent",
			text:
`/*
 * some text
 *   line two
 * line three
*/`,
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: "\n * ", Indent: " * ", Suffix: "\n",
				Text: "some text\n  line two\nline three",
			},
		},
		{
			name: "inconsistent",
			text:
`/*
 * some text
   line two
 * line three
*/`,
			exp: commentElems{
				StartToken: "/*", EndToken: "*/",
				Prefix: "\n * ", Indent: " ", Suffix: "\n",
				Text: "some text\n  line two\n* line three",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := commentElems{
				StartToken: c.exp.StartToken,
				EndToken:   c.exp.EndToken,
			}
			if !v.Split(c.text) {
				t.Error("split failed")
			}
			require.Equal(t, c.exp, v)
			require.Equal(t, c.text, v.Join())
		})
	}
}
