package transformer

import (
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func uastType(uobj interface{}, op ObjectOp, part string) ObjectOp {
	if op == nil {
		op = Obj{}
	}
	utyp := uast.TypeOf(uobj)
	if utyp == "" {
		panic(fmt.Errorf("type is not registered: %T", uobj))
	}
	obj := Obj{uast.KeyType: String(utyp)}
	if part != "" {
		return JoinObj(obj, Part(part, op))
	}
	fields, ok := op.Fields()
	if !ok {
		return JoinObj(obj, op)
	}
	zero, opt := uast.NewObjectByTypeOpt(utyp)
	delete(zero, uast.KeyType)
	if len(zero) == 0 {
		return JoinObj(obj, op)
	}
	for k := range fields {
		if k == uast.KeyType {
			continue
		}
		_, ok := zero[k]
		_, ok2 := opt[k]
		if !ok && !ok2 {
			panic(ErrUndefinedField.New(utyp + "." + k))
		}
		delete(zero, k)
	}
	for k, v := range zero {
		obj[k] = Is(v)
	}
	return JoinObj(obj, op)
}

func UASTType(uobj interface{}, op ObjectOp) ObjectOp {
	return uastType(uobj, op, "")
}

func UASTTypePart(vr string, uobj interface{}, op ObjectOp) ObjectOp {
	return uastType(uobj, op, vr)
}

func remapPos(m ObjMapping, names map[string]string) ObjMapping {
	so, do := m.ObjMapping() // TODO: clone?

	sp := UASTType(uast.Positions{}, Fields{
		{Name: uast.KeyStart, Op: Var(uast.KeyStart), Optional: uast.KeyStart + "_exists"},
		{Name: uast.KeyEnd, Op: Var(uast.KeyEnd), Optional: uast.KeyEnd + "_exists"},
	})
	dp := UASTType(uast.Positions{}, Fields{
		{Name: uast.KeyStart, Op: Var(uast.KeyStart), Optional: uast.KeyStart + "_exists"},
		{Name: uast.KeyEnd, Op: Var(uast.KeyEnd), Optional: uast.KeyEnd + "_exists"},
	})
	if len(names) != 0 {
		sa, da := make(Obj), make(Obj)
		for k, v := range names {
			sa[k] = Var(v)
			if v != uast.KeyStart && v != uast.KeyEnd {
				da[k] = Var(v)
			}
		}
		sp, dp = JoinObj(sp, sa), JoinObj(dp, da)
	}
	return MapObj(
		JoinObj(so, Obj{uast.KeyPos: sp}),
		JoinObj(do, Obj{uast.KeyPos: dp}),
	)
}

func MapSemantic(nativeType string, semType interface{}, m ObjMapping) ObjMapping {
	return MapSemanticPos(nativeType, semType, nil, m)
}

func MapSemanticPos(nativeType string, semType interface{}, pos map[string]string, m ObjMapping) ObjMapping {
	so, do := m.ObjMapping() // TODO: clone?
	so = JoinObj(Obj{uast.KeyType: String(nativeType)}, so)
	so, do = remapPos(MapObj(so, do), pos).ObjMapping()
	return MapObj(so, UASTType(semType, do))
}

func CommentText(tokens [2]string, vr string) Op {
	return &commentUAST{
		tokens: tokens,
		text:   vr + "_text", pref: vr + "_pref", suff: vr + "_suff", tab: vr + "_tab",
	}
}

func CommentNode(block bool, vr string, pos Op) ObjectOp {
	obj := Obj{
		"Block":  Bool(block),
		"Text":   Var(vr + "_text"),
		"Prefix": Var(vr + "_pref"),
		"Suffix": Var(vr + "_suff"),
		"Tab":    Var(vr + "_tab"),
	}
	if pos != nil {
		obj[uast.KeyPos] = pos
	}
	return UASTType(uast.Comment{}, obj)
}

// commentElems contains individual comment elements.
// See uast.Comment for details.
type commentElems struct {
	Tokens [2]string
	Text   string
	Pref   string
	Suff   string
	Tab    string
}

func (c *commentElems) isTab(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	for _, r2 := range c.Tokens[0] {
		if r == r2 {
			return true
		}
	}
	for _, r2 := range c.Tokens[1] {
		if r == r2 {
			return true
		}
	}
	return false
}

func (c *commentElems) Split(text string) bool {
	if !strings.HasPrefix(text, c.Tokens[0]) || !strings.HasSuffix(text, c.Tokens[1]) {
		return false
	}
	text = strings.TrimPrefix(text, c.Tokens[0])
	text = strings.TrimSuffix(text, c.Tokens[1])

	// find prefix
	i := strings.IndexFunc(text, func(r rune) bool {
		return !c.isTab(r)
	})
	c.Pref = text[:i]
	text = text[i:]

	// find suffix
	i = strings.LastIndexFunc(text, func(r rune) bool {
		return !c.isTab(r)
	})
	c.Suff = text[i+1:]
	text = text[:i+1]
	c.Text = text

	sub := strings.Split(text, "\n")
	if len(sub) == 1 {
		// fast path, no tabs
		return true
	}

	// find minimal common prefix for other lines
	// first line is special, it won't contain tab
	for i, line := range sub[1:] {
		if i == 0 {
			j := strings.IndexFunc(line, func(r rune) bool {
				return !c.isTab(r)
			})
			c.Tab = line[:j]
			if c.Tab == "" {
				return true // no tabs
			}
			continue
		}
		if strings.HasPrefix(line, c.Tab) {
			continue
		}
		j := strings.IndexFunc(line, func(r rune) bool {
			return !c.isTab(r)
		})
		tab := line[:j]
		if tab == "" {
			return true // no tabs
		}
		for j := 0; j < len(c.Tab) && j < len(tab); j++ {
			if c.Tab[j] == tab[j] {
				continue
			}
			if j == 0 {
				return true // inconsistent, no tabs
			}
			tab = tab[:j]
			break
		}
		c.Tab = tab
	}
	for i, line := range sub {
		if i == 0 {
			continue
		}
		sub[i] = strings.TrimPrefix(line, c.Tab)
	}
	c.Text = strings.Join(sub, "\n")
	return true
}

func (c commentElems) Join() string {
	if c.Tab != "" {
		sub := strings.Split(c.Text, "\n")
		for i, line := range sub {
			if i == 0 {
				continue
			}
			sub[i] = c.Tab + line
		}
		c.Text = strings.Join(sub, "\n")
	}
	return strings.Join([]string{
		c.Tokens[0], c.Pref,
		c.Text,
		c.Suff, c.Tokens[1],
	}, "")
}

type commentUAST struct {
	tokens          [2]string
	text            string
	pref, suff, tab string
}

func (*commentUAST) Kinds() nodes.Kind {
	return nodes.KindString
}

func (op *commentUAST) Check(st *State, n nodes.Node) (bool, error) {
	s, ok := n.(nodes.String)
	if !ok {
		return false, nil
	}

	c := commentElems{Tokens: op.tokens}
	if !c.Split(string(s)) {
		return false, nil
	}

	err := st.SetVars(Vars{
		op.text: nodes.String(c.Text),
		op.pref: nodes.String(c.Pref),
		op.suff: nodes.String(c.Suff),
		op.tab:  nodes.String(c.Tab),
	})
	return err == nil, err
}

func (op *commentUAST) Construct(st *State, n nodes.Node) (nodes.Node, error) {
	var (
		text, pref, suff, tab nodes.String
	)

	err := st.MustGetVars(VarsPtrs{
		op.text: &text,
		op.pref: &pref, op.suff: &suff, op.tab: &tab,
	})
	if err != nil {
		return nil, err
	}

	c := commentElems{
		Tokens: op.tokens,
		Text:   string(text),
		Pref:   string(pref),
		Suff:   string(suff),
		Tab:    string(tab),
	}

	return nodes.String(c.Join()), nil
}
