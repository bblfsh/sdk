package positioner

import (
	"fmt"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
)

var _ transformer.CodeTransformer = TokenFromSource{}

// TokenFromSource extract node's token from the source code by using positional
// information.
type TokenFromSource struct {
	// Key is the name of the token field to update. Uses uast.KeyToken, if not set.
	// Only nodes with this field will be considered.
	Key string
	// Types is the list of node types that will be updated. Empty means all nodes.
	Types []string
}

// OnCode implements transformer.CodeTransformer.
func (t TokenFromSource) OnCode(code string) transformer.Transformer {
	return &tokenFromSource{
		tokenFilter: newTokenFilter(code, t.Key, t.Types),
	}
}

func newTokenFilter(source string, key string, types []string) tokenFilter {
	f := tokenFilter{
		source: source, tokenKey: key,
	}
	if f.tokenKey == "" {
		f.tokenKey = uast.KeyToken
	}
	if len(types) != 0 {
		f.types = make(map[string]struct{})
		for _, tp := range types {
			f.types[tp] = struct{}{}
		}
	}
	return f
}

type tokenFilter struct {
	source   string
	tokenKey string
	types    map[string]struct{}
}

func (f *tokenFilter) filterObj(node nodes.Node) (nodes.Object, bool) {
	obj, ok := node.(nodes.Object)
	if !ok {
		return nil, false
	}
	// nodes should already have a token
	if _, ok = obj[f.tokenKey]; !ok {
		return nil, false
	}
	// apply types filter
	if f.types != nil {
		typ, ok := obj[uast.KeyType].(nodes.String)
		if !ok || typ == "" {
			return nil, false
		}
		_, ok = f.types[string(typ)]
		if !ok {
			return nil, false
		}
	}
	return obj, true
}

func (f *tokenFilter) tokenFromPos(obj nodes.Object) (string, bool, error) {
	pos := uast.PositionsOf(obj)
	if len(pos) == 0 {
		return "", false, nil
	}
	start := pos.Start()
	end := pos.End()
	if start == nil || end == nil || !start.HasOffset() || !end.HasOffset() {
		return "", false, nil
	}
	si, ei := start.Offset, end.Offset
	if si > ei {
		return "", false, fmt.Errorf("start offset is larger than an end offset: %v", pos)
	} else if ei > uint32(len(f.source)) {
		return "", false, fmt.Errorf("offset out of bounds: %v", pos)
	}
	// offset is given in bytes
	token := f.source[si:ei]
	return token, true, nil
}

type tokenFromSource struct {
	tokenFilter
}

// Do implements transformer.Transformer. See TokenFromSource.
func (t *tokenFromSource) Do(root nodes.Node) (nodes.Node, error) {
	var last error
	nodes.WalkPreOrder(root, func(node nodes.Node) bool {
		if last != nil {
			return false
		}
		obj, ok := t.filterObj(node)
		if !ok {
			// skip node, but recurse to children
			return true
		}
		token, ok, err := t.tokenFromPos(obj)
		if err != nil {
			last = err
			return false
		} else if !ok {
			return true // recurse
		}
		// it won't be nil, since we require both token and pos fields to exist
		obj[t.tokenKey] = nodes.String(token)
		return true
	})
	return root, last
}

// VerifyToken check that node's token matches its positional information.
type VerifyToken struct {
	// Key is the name of the token field to check. Uses uast.KeyToken, if not set.
	Key string
	// Types is the list of node types that will be checked. Empty means all nodes.
	Types []string
}

func (t VerifyToken) Verify(code string, root nodes.Node) error {
	key := t.Key
	if key == "" {
		key = uast.KeyToken
	}
	f := newTokenFilter(code, key, t.Types)

	var last error
	nodes.WalkPreOrder(root, func(node nodes.Node) bool {
		if last != nil {
			return false
		}
		obj, ok := f.filterObj(node)
		if !ok {
			// skip node, but recurse to children
			return true
		}
		token1, ok := obj[key].(nodes.String)
		if !ok {
			return true
		}
		token2, ok, err := f.tokenFromPos(obj)
		if err != nil {
			last = err
			return false
		} else if !ok {
			return true
		}
		if string(token1) != token2 {
			last = fmt.Errorf("wrong token for node %q: %q vs %q",
				uast.TypeOf(obj), token1, token2)
			return false
		}
		return true
	})
	return last
}
