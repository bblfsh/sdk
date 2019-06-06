package uast

import (
	"sort"
	"strings"

	"github.com/bblfsh/sdk/v3/uast/nodes"
)

// AllImportPaths returns a list of all import paths in the UAST. Resulting import paths will be deduplicated and sorted.
//
// Path elements in QualifiedIdentifiers import will be joined by '/'. For example, Java import "com.example.pkg"
// will be listed as "com/example/pkg".
func AllImportPaths(root nodes.External) []string {
	it := nodes.NewIterator(root, nodes.IterAny)
	var (
		paths []string
		seen  = make(map[string]struct{})
	)
	for it.Next() {
		n := it.Node()
		if n.Kind() != nodes.KindObject {
			continue
		}
		path, ok := getImportPath(n)
		if !ok || path == "" {
			continue
		} else if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// getImportPath returns a concatenated import path of any node derived from Import and returns false if conversion fails.
// See AllImportPaths for details.
//
// This function won't decode the Import and won't verify the schema, except for the Path field.
func getImportPath(imp nodes.External) (string, bool) {
	typ := TypeOf(imp)
	if !strings.HasPrefix(typ, NS+":") {
		return "", false
	}
	switch typ {
	default:
		return "", false
	case importType, runtimeImportType, runtimeReImportType, inlineImportType: // TODO: automatically build this list using reflection
	}
	var _ Import // helps to find Import usages in IDE
	path, ok := getField(imp, "Path", nodes.KindObject)
	if !ok || path == nil {
		return "", false
	}
	return asImportPath(path)
}

// asImportPath returns a concatenated import path of any node suitable for Import.Path field.
func asImportPath(path nodes.External) (string, bool) {
	switch TypeOf(path) {
	case stringType:
		return getStringValue(path)
	case identifierType:
		return getIdentifierName(path)
	case qualifiedIdentifierType:
		names, ok := getQualifiedIdentifierNames(path)
		if !ok {
			return "", false
		}
		return strings.Join(names, "/"), true
	case aliasType:
		n, ok := getAliasNode(path)
		if !ok {
			return "", false
		}
		return asImportPath(n)
	default:
		return "", false
	}
}

// getField extracts a specified field from the node.
func getField(n nodes.External, key string, kind nodes.Kind) (nodes.External, bool) {
	if n == nil || n.Kind() != nodes.KindObject {
		return nil, false
	}
	switch obj := n.(type) {
	case nodes.Object:
		v, ok := obj[key]
		if !ok || v == nil || !v.Kind().In(kind) {
			return nil, false
		}
		return v, ok
	case nodes.ExternalObject:
		v, ok := obj.ValueAt(key)
		if !ok || v == nil || !v.Kind().In(kind) {
			return nil, false
		}
		return v, ok
	default:
		return nil, false
	}
}

// getValueField extracts a specified value field from the node.
func getValueField(n nodes.External, key string, kind nodes.Kind) (nodes.Value, bool) {
	v, ok := getField(n, key, kind)
	if !ok {
		return nil, false
	}
	if v, ok := v.(nodes.Value); ok {
		return v, true
	}
	return v.Value(), true
}

// getStringField extracts a specified string field from the node.
func getStringField(n nodes.External, key string) (string, bool) {
	v, ok := getValueField(n, key, nodes.KindString)
	if !ok {
		return "", false
	}
	s, ok := v.(nodes.String)
	return string(s), ok
}

// getStringValue extracts the String.Value field from the node.
func getStringValue(n nodes.External) (string, bool) {
	var _ String // helps to find String usages in IDE
	if TypeOf(n) != stringType {
		return "", false
	}
	return getStringField(n, "Value")
}

// getIdentifierName extracts the Identifier.Name field from the node.
func getIdentifierName(n nodes.External) (string, bool) {
	var _ Identifier // helps to find Identifier usages in IDE
	if TypeOf(n) != identifierType {
		return "", false
	}
	return getStringField(n, "Name")
}

// getQualifiedIdentifierNames extracts the QualifiedIdentifier.Names field from the node.
func getQualifiedIdentifierNames(n nodes.External) ([]string, bool) {
	var _ QualifiedIdentifier // may help to find usages
	if TypeOf(n) != qualifiedIdentifierType {
		return nil, false
	}
	arr, ok := getField(n, "Names", nodes.KindArray)
	if !ok || arr == nil {
		return nil, false
	}
	switch arr := arr.(type) {
	case nodes.Array:
		names := make([]string, 0, len(arr))
		for _, v := range arr {
			name, ok := getIdentifierName(v)
			if !ok {
				return nil, false
			}
			names = append(names, name)
		}
		return names, len(names) != 0
	case nodes.ExternalArray:
		sz := arr.Size()
		names := make([]string, 0, sz)
		for i := 0; i < sz; i++ {
			v := arr.ValueAt(i)
			if v == nil || TypeOf(v) != identifierType {
				return nil, false
			}
			name, ok := getIdentifierName(v)
			if !ok {
				return nil, false
			}
			names = append(names, name)
		}
		return names, len(names) != 0
	default:
		return nil, false
	}
}

// getAliasNode extracts the Alias.Node field from the node.
func getAliasNode(n nodes.External) (nodes.External, bool) {
	var _ Alias // helps to find Alias usages in IDE
	if TypeOf(n) != aliasType {
		return nil, false
	}
	return getField(n, "Node", nodes.KindObject)
}
