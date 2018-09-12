package xpath

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v2/uast/role"

	"github.com/stretchr/testify/require"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/query"
)

func mustNode(o interface{}) nodes.Node {
	n, err := uast.ToNode(o)
	if err != nil {
		panic(err)
	}
	return n
}

func TestFilter(t *testing.T) {
	var root = nodes.Array{
		mustNode(uast.Identifier{Name: "Foo"}),
		nodes.Object{
			uast.KeyType:  nodes.String("Ident"),
			uast.KeyToken: nodes.String("A"),
			uast.KeyRoles: nodes.Array{
				nodes.Int(role.Identifier),
				nodes.Int(role.Name),
			},
		},
	}

	idx := New()

	it, err := idx.Execute(root, "//uast:Identifier[@Name='Foo']")
	require.NoError(t, err)
	expect(t, it, root[0])

	it, err = idx.Execute(root, "//uast:Identifier/Name[text() = 'Foo']/..")
	require.NoError(t, err)
	expect(t, it, root[0])

	it, err = idx.Execute(root, "//Identifier")
	require.NoError(t, err)
	expect(t, it)

	it, err = idx.Execute(root, "//Ident")
	require.NoError(t, err)
	expect(t, it, root[1])

	it, err = idx.Execute(root, "//Ident[text() = 'A']")
	require.NoError(t, err)
	expect(t, it, root[1])

	it, err = idx.Execute(root, "//Ident[@role = 'Name']")
	require.NoError(t, err)
	expect(t, it, root[1])

	it, err = idx.Execute(root, "//Ident[@role = 'Invalid']")
	require.NoError(t, err)
	expect(t, it)
}

func TestFilterObject(t *testing.T) {
	b := nodes.Object{
		uast.KeyType: nodes.String("B"),
	}
	c := nodes.Object{
		uast.KeyType: nodes.String("C"),
	}
	d := nodes.Object{
		uast.KeyType: nodes.String("d:X"),
	}
	v := nodes.String("val")
	va, vb := nodes.String("a"), nodes.String("b")
	varr := nodes.Array{
		va,
		vb,
	}
	var root = nodes.Object{
		uast.KeyType: nodes.String("A"),
		"key":        v,
		"keys":       varr,
		"one":        b,
		"sub": nodes.Array{
			c,
			d,
		},
	}
	/*
		<A key='val' keys='a' keys='b'>
			<key>val</key>
			<keys>a</keys>
			<keys>b</keys>
			<one>
				<B></B>
			</one>
			<sub>
				<C></C>
				<d:X></d:X>
			</sub>
		</A>
	*/

	idx := New()

	queries := []struct {
		name string
		qu   string
		exp  []nodes.Node
	}{
		{
			name: "root", qu: "/",
			exp: []nodes.Node{root},
		},
		{
			name: "root tag", qu: "/A",
			exp: []nodes.Node{root},
		},
		{
			name: "field obj", qu: "/A/one",
			exp: []nodes.Node{b},
		},
		{
			name: "field obj tag", qu: "/A/one/B",
			exp: []nodes.Node{b},
		},
		{
			name: "field obj arr", qu: "/A/sub",
			exp: []nodes.Node{nodes.Array{c, d}},
		},
		{
			name: "field obj arr tag", qu: "/A/sub/C",
			exp: []nodes.Node{c},
		},
		{
			name: "inner field", qu: "//one",
			exp: []nodes.Node{b},
		},
		{
			name: "inner obj", qu: "//B",
			exp: []nodes.Node{b},
		},
		{
			name: "inner obj 2", qu: "//C",
			exp: []nodes.Node{c},
		},
		{
			name: "inner obj ns", qu: "//d:X",
			exp: []nodes.Node{d},
		},
		{
			name: "field value", qu: "/A/key",
			exp: []nodes.Node{v},
		},
		{
			name: "field value text", qu: "/A/key[text() = 'val']",
			exp: []nodes.Node{v},
		},
		{
			name: "field value arr", qu: "/A/keys",
			exp: []nodes.Node{varr},
		},
		{
			name: "text", qu: "//*[text() = 'a']",
			exp: []nodes.Node{varr},
		},
		{
			name: "attr value", qu: "//A[@key='val']",
			exp: []nodes.Node{root},
		},
		{
			name: "attr value arr", qu: "//A[@keys='a']",
			exp: []nodes.Node{root},
		},
		// TODO: fix in xpath library
		//{
		//	name: "field value arr elem", qu: "/A/keys/",
		//	exp: []nodes.Node{va, vb},
		//},
	}

	for _, c := range queries {
		c := c
		t.Run(c.name, func(t *testing.T) {
			it, err := idx.Execute(root, c.qu)
			require.NoError(t, err)
			expect(t, it, c.exp...)
		})
	}
}

func expect(t testing.TB, it query.Iterator, exp ...nodes.Node) {
	var out []nodes.Node
	for it.Next() {
		out = append(out, it.Node().(nodes.Node))
	}
	require.Equal(t, exp, out)
}
