package uast

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	ErrIncorrectType = errors.NewKind("incorrect object type: %q, expected: %q")
)

var (
	namespaces = make(map[string]string)
	package2ns = make(map[string]string)
)

func RegisterPackage(ns string, o interface{}) {
	if _, ok := namespaces[ns]; ok {
		panic("namespace already registered")
	}
	pkg := reflect.TypeOf(o).PkgPath()
	if _, ok := package2ns[pkg]; ok {
		panic("package already registered")
	}
	namespaces[ns] = pkg
	package2ns[pkg] = ns
}

func TypeOf(o interface{}) string {
	if o == nil {
		return ""
	} else if obj, ok := o.(nodes.Object); ok {
		tp, _ := obj[KeyType].(nodes.String)
		return string(tp)
	}
	tp := reflect.TypeOf(o)
	ns, name := typeOf(tp)
	if ns == "" {
		return name
	}
	return ns + ":" + name
}

func typeOf(tp reflect.Type) (ns, name string) {
	pkg := tp.PkgPath()
	if pkg == "" {
		return
	}
	name = tp.Name()
	if name == "" {
		return
	}
	ns = package2ns[pkg]
	return
}

func fieldName(f reflect.StructField) (string, error) {
	name := strings.SplitN(f.Tag.Get("uast"), ",", 2)[0]
	if name == "" {
		name = strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
	}
	if name == "" {
		return "", fmt.Errorf("field %s should have uast or json name", f.Name)
	}
	return name, nil
}

var (
	reflString = reflect.TypeOf("")
	reflNode   = reflect.TypeOf((*nodes.Node)(nil)).Elem()
)

// ToNode converts objects returned by schema-less encodings such as JSON to Node objects.
// It also supports types from packages registered via RegisterPackage.
func ToNode(o interface{}) (nodes.Node, error) {
	return nodes.ToNode(o, func(o interface{}) (nodes.Node, error) {
		return toNodeReflect(reflect.ValueOf(o))
	})
}

func toNodeReflect(rv reflect.Value) (nodes.Node, error) {
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rv.Type()
	}
	switch rt.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return nodes.Int(rv.Int()), nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return nodes.Int(rv.Uint()), nil
	case reflect.Float64, reflect.Float32:
		return nodes.Float(rv.Float()), nil
	case reflect.Bool:
		return nodes.Bool(rv.Bool()), nil
	case reflect.String:
		return nodes.String(rv.String()), nil
	case reflect.Slice:
		// TODO: catch []byte
		arr := make(nodes.Array, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			v, err := toNodeReflect(rv.Index(i))
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	case reflect.Struct, reflect.Map:
		ns, name := typeOf(rt)
		if ns == "" {
			return nil, fmt.Errorf("type %v is not registered", rt)
		}
		typ := ns + ":" + name

		isStruct := rt.Kind() == reflect.Struct

		sz := 0
		if isStruct {
			sz = rt.NumField()
		} else {
			sz = rv.Len()
		}

		obj := make(nodes.Object, sz+1)
		obj[KeyType] = nodes.String(typ)

		if isStruct {
			for i := 0; i < rt.NumField(); i++ {
				f := rv.Field(i)
				if !f.CanInterface() {
					continue
				}
				ft := rt.Field(i)
				name, err := fieldName(ft)
				if err != nil {
					return nil, fmt.Errorf("type %s: %v", rt.Name(), err)
				}
				v, err := toNodeReflect(f)
				if err != nil {
					return nil, err
				}
				obj[name] = v
			}
		} else {
			if rt.Key() != reflString {
				return nil, fmt.Errorf("unsupported map key type: %v", rt.Key())
			}
			for _, k := range rv.MapKeys() {
				v, err := toNodeReflect(rv.MapIndex(k))
				if err != nil {
					return nil, err
				}
				obj[k.String()] = v
			}
		}
		return obj, nil
	}
	return nil, fmt.Errorf("unsupported type: %v", rt)
}

func NodeAs(n nodes.Node, dst interface{}) error {
	var rv reflect.Value
	if v, ok := dst.(reflect.Value); ok {
		rv = v
	} else {
		rv = reflect.ValueOf(dst)
	}
	return nodeAs(n, rv)
}

func nodeAs(n nodes.Node, rv reflect.Value) error {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.CanSet() {
		return fmt.Errorf("argument should be a pointer: %v", rv.Type())
	}
	switch n := n.(type) {
	case nil:
		return nil
	case nodes.Object:
		rt := rv.Type()
		kind := rt.Kind()
		if kind != reflect.Struct && kind != reflect.Map {
			return fmt.Errorf("expected struct or map, got %v", rt)
		}
		ns, name := typeOf(rt)
		etyp := ns + ":" + name
		if typ := TypeOf(n); typ != etyp {
			return ErrIncorrectType.New(typ, etyp)
		}
		if kind == reflect.Struct {
			for i := 0; i < rt.NumField(); i++ {
				f := rv.Field(i)
				if !f.CanInterface() {
					continue
				}
				ft := rt.Field(i)
				name, err := fieldName(ft)
				if err != nil {
					return fmt.Errorf("type %s: %v", rt.Name(), err)
				}
				v, ok := n[name]
				if !ok {
					continue
				}
				if err = nodeAs(v, f); err != nil {
					return err
				}
			}
		} else {
			if rv.IsNil() {
				rv.Set(reflect.MakeMapWithSize(rt, len(n)-1))
			}
			for k, v := range n {
				if k == KeyType {
					continue
				}
				nv := reflect.New(rt.Elem()).Elem()
				if err := nodeAs(v, nv); err != nil {
					return err
				}
				rv.SetMapIndex(reflect.ValueOf(k), nv)
			}
		}
		return nil
	case nodes.Array:
		rt := rv.Type()
		if rt.Kind() != reflect.Slice {
			return fmt.Errorf("expected slice, got %v", rt)
		}
		if rv.Cap() < len(n) {
			rv.Set(reflect.MakeSlice(rt.Elem(), len(n), len(n)))
		} else {
			rv = rv.Slice(0, len(n))
		}
		for i, v := range n {
			if err := nodeAs(v, rv.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case nodes.String, nodes.Int, nodes.Float, nodes.Bool:
		rt := rv.Type()
		nv := reflect.ValueOf(n)
		if !nv.Type().ConvertibleTo(rt) {
			return fmt.Errorf("cannot convert %T to %v", n, rt)
		}
		rv.Set(nv.Convert(rt))
		return nil
	}
	return fmt.Errorf("unexpected type: %T", n)
}
