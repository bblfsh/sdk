package diff

import (
	"fmt"

	"github.com/heetch/lapjv"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type decision interface {
	cost() int
}

type same struct {
	c int
}
type replace struct {
	c int
}
type match struct {
	c int
}
type permute struct {
	c           int
	permutation []int
}

func (d same) cost() int    { return d.c }
func (d replace) cost() int { return d.c }
func (d match) cost() int   { return d.c }
func (d permute) cost() int { return d.c }

// min is a convenience method for choosing the cheapest decision
func min(self, candidate decision) decision {
	if self.cost() > candidate.cost() {
		return candidate
	}
	return self
}

// keyType for cache
type keyType struct{ k1, k2 ID }

// cacheStorage is a cache for diff computation
type cacheStorage struct {
	lastGroup uint64
	decisions map[keyType]decision
	counts    map[ID]int
	changes   Changelist
}

func (ds *cacheStorage) newGroup() uint64 {
	ds.lastGroup++
	return ds.lastGroup
}

func makeCacheStorage() *cacheStorage {
	return &cacheStorage{
		decisions: make(map[keyType]decision),
		counts:    make(map[ID]int),
	}
}

// nodeSize caches (for perf) tree size (node count)
func (ds *cacheStorage) nodeSize(n nodes.Node) int {
	label := nodes.UniqueKey(n)
	if cnt, ok := ds.counts[label]; ok {
		return cnt
	}
	ret := nodes.Count(n, nodes.KindsNotNil)
	ds.counts[label] = ret
	return ret
}

// decideAction finds the cheapest way to naively match src and dst and returns this action
// with its combined cost
func (ds *cacheStorage) decideAction(src, dst nodes.Node) decision {
	label := keyType{nodes.UniqueKey(src), nodes.UniqueKey(dst)}

	if val, ok := ds.decisions[label]; ok {
		return val
	}

	// the default (but not always optimal) decision is to ignore the existance of src and
	// just create the desired dst; implemented below
	cost := ds.nodeSize(dst) + 1

	var best decision = replace{cost}
	if nodes.KindOf(src) != nodes.KindOf(dst) {
		ds.decisions[label] = best
		return best
	}

	// from now on src.(type) == dst.(type)

	switch src := src.(type) {
	case nil:
		cost = 0
		best = min(best, same{cost})
	case nodes.Value:
		dst := dst.(nodes.Value)
		cost = 0
		if src == dst {
			best = min(best, same{cost})
		} else {
			cost = 1
			best = min(best, replace{cost})
		}
	case nodes.Object:
		dst := dst.(nodes.Object)
		cost = 0

		// the code below iterates over each of the keys from src and dst exactly once,
		// that's why keys isn't reset between iterations.
		for _, key := range src.Keys() {
			cost += ds.decideAction(src[key], dst[key]).cost()
		}
		for _, key := range dst.Keys() {
			if _, in := src[key]; !in {
				cost += ds.decideAction(src[key], dst[key]).cost()
			}
		}

		if cost == 0 {
			best = min(best, same{cost})
		} else {
			best = min(best, match{cost})
		}
	case nodes.Array:
		dst := dst.(nodes.Array)
		cost = 0
		if len(src) == len(dst) {
			sum := 0
			for i := range src {
				sum += ds.decideAction(src[i], dst[i]).cost()
			}
			if sum == 0 {
				best = min(best, same{cost})
				break
			}
		} else {
			cost = 2
		}

		if len(src) < len(dst) {
			arr := make(nodes.Array, len(dst))
			copy(arr, src)
			src = arr
		} else if len(src) > len(dst) {
			arr := make(nodes.Array, len(src))
			copy(arr, dst)
			dst = arr
		}
		n := len(src)
		m := make([][]int, n)
		for i := range src {
			m[i] = make([]int, n)
			for j := range dst {
				m[i][j] = ds.decideAction(src[i], dst[j]).cost()
			}
		}

		res := lapjv.Lapjv(m)

		for i1, i2 := range res.InRow {
			if i1 != i2 {
				cost = 2
				break
			}
		}

		cost += res.Cost
		best = min(best, permute{cost, res.InRow})
	default:
		panic(fmt.Errorf("unknown node type %T", src))
	}

	ds.decisions[label] = best
	return best
}

func (ds *cacheStorage) push(change Change) {
	ds.changes = append(ds.changes, change)
}

func (ds *cacheStorage) createRec(group uint64, node nodes.Node) {
	switch n := node.(type) {
	case nodes.Object:
		for _, v := range n {
			ds.createRec(group, v)
		}
	case nodes.Array:
		for _, v := range n {
			ds.createRec(group, v)
		}
	default:
		// values and nils are not saved separately
		return
	}
	ds.push(Create{group: group, Node: node})
}

func (ds *cacheStorage) generateDifference(src, dst nodes.Node, parentID ID, parentKey Key) {
	switch d := ds.decideAction(src, dst).(type) {
	case same:
	// no action required if same
	case replace:
		g := ds.newGroup()
		// remove src (no action?) and create dst + attach it
		if dst != nil {
			ds.createRec(g, dst)
			ds.push(Attach{
				group:  g,
				Parent: parentID,
				Key:    parentKey,
				Child:  nodes.UniqueKey(dst),
			})
		} else {
			ds.push(Attach{
				group:  g,
				Parent: parentID,
				Key:    parentKey,
				Child:  nil,
			})
		}
	case match:
		src, dst := src.(nodes.Object), dst.(nodes.Object)

		checkAndPush := func(key string) {
			if _, ok := dst[key]; !ok {
				ds.push(Detach{
					Parent: nodes.UniqueKey(src),
					Key:    String(key),
				})
			} else if _, ok := src[key]; !ok {
				g := ds.newGroup()
				ds.createRec(g, dst[key])
				ds.push(Attach{
					group:  g,
					Parent: nodes.UniqueKey(src),
					Key:    String(key),
					Child:  nodes.UniqueKey(dst[key]),
				})
			} else {
				ds.generateDifference(
					src[key], dst[key], nodes.UniqueKey(src), String(key))
			}
		}
		for _, key := range src.Keys() {
			checkAndPush(key)
		}
		for _, key := range dst.Keys() {
			if _, in := src[key]; !in {
				checkAndPush(key)
			}
		}
	case permute:
		src, dst := src.(nodes.Array), dst.(nodes.Array)
		l := len(dst) - len(src)
		// add possible nils to src
		if l > 0 {
			arr := make(nodes.Array, len(dst))
			copy(arr, src)
			src = arr
		}
		recreate := l != 0
		for i1, i2 := range d.permutation {
			if i1 != i2 {
				recreate = true
				break
			}
		}
		if recreate {
			// recreate src with right perm
			newsrc := make([]nodes.Node, 0, len(dst))
			for i := range dst {
				newsrc = append(newsrc, src[d.permutation[i]])
			}
			src = newsrc
			g := ds.newGroup()
			ds.push(Create{group: g, Node: src}) // TODO: not create, only mutate...
			ds.push(Attach{
				group:  g,
				Parent: parentID,
				Key:    parentKey,
				Child:  nodes.UniqueKey(src),
			})
		}
		for i := 0; i < len(dst); i++ {
			ds.generateDifference(src[i], dst[i], nodes.UniqueKey(src), Int(i))
		}
	default:
		panic(fmt.Errorf("unknown decision %v", d))
	}
}

// Cost is a function that takes two trees: src and dst and returns number of operations
// needed to convert the former into the latter.
func Cost(src, dst nodes.Node) int {
	ds := makeCacheStorage()
	return ds.decideAction(src, dst).cost()
}

// Changes is a function that takes two trees: src and dst and returns a changelist
// containing all operations required to convert the former into the latter.
func Changes(src, dst nodes.Node) Changelist {
	ds := makeCacheStorage()
	ds.generateDifference(src, dst, nil, Int(0))
	return ds.changes
}
