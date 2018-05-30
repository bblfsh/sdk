package transformer

import (
	"fmt"
	"strings"
	"unicode"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func UASTType(uobj interface{}, op ObjectOp) ObjectOp {
	utyp := uast.TypeOf(uobj)
	if utyp == "" {
		panic(fmt.Errorf("type is not registered: %T", uobj))
	}
	obj := op.Object()
	obj.SetField(uast.KeyType, String(utyp))
	return obj
}

func MapSemantic(name, nativeType string, semType interface{}, pos map[string]string, src, dst ObjectOp) Mapping {
	utyp := uast.TypeOf(semType)
	if strings.HasPrefix(name, " ") {
		name = nativeType + " -> " + utyp + name
	} else if name == "" {
		name = nativeType + " -> " + utyp
	}
	so, do := src.Object(), dst.Object()

	sp := UASTType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	dp := UASTType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	for k, v := range pos {
		sp.SetField(k, Var(v))
		if v != "start" && v != "end" {
			dp.SetField(k, Var(v))
		}
	}
	so.SetField(uast.KeyType, String(nativeType))
	so.SetField(uast.KeyPos, sp)
	do.SetField(uast.KeyPos, dp)
	return Map(name, so, UASTType(semType, do))
}

func CommentText(tokens [2]string, vr string) Op {
	return commentUAST{
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

type commentUAST struct {
	tokens          [2]string
	text            string
	pref, suff, tab string
}

func (commentUAST) Kinds() nodes.Kind {
	return nodes.KindString
}

func (op commentUAST) Check(st *State, n nodes.Node) (bool, error) {
	s, ok := n.(nodes.String)
	if !ok {
		return false, nil
	}
	text := string(s)
	if !strings.HasPrefix(text, op.tokens[0]) || !strings.HasSuffix(text, op.tokens[1]) {
		return false, nil
	}
	text = strings.TrimPrefix(text, op.tokens[0])
	text = strings.TrimSuffix(text, op.tokens[1])
	var (
		pref, suff, tab string
	)

	// find prefix
	i := 0
	for ; i < len(text); i++ {
		if r := rune(text[i]); unicode.IsLetter(r) || unicode.IsNumber(r) {
			break
		}
	}
	pref = text[:i]
	text = text[i:]

	// find suffix
	i = len(text) - 1
	for ; i >= 0 && unicode.IsSpace(rune(text[i])); i-- {
	}
	suff = text[i+1:]
	text = text[:i+1]

	// TODO: set tab

	err := st.SetVars(Vars{
		op.text: nodes.String(text),
		op.pref: nodes.String(pref),
		op.suff: nodes.String(suff),
		op.tab:  nodes.String(tab),
	})
	return err == nil, err
}

func (op commentUAST) Construct(st *State, n nodes.Node) (nodes.Node, error) {
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
	// FIXME: handle tab
	text = pref + text + suff
	return nodes.String(op.tokens[0] + string(text) + op.tokens[1]), nil
}
