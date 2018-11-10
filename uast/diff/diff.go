package diff

import (
	"fmt"
	"github.com/heetch/lapjv"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type decisionType interface {
	cost() int
}

type basicDecision struct {
	privateCost int
}

func (b basicDecision) cost() int { return b.privateCost }

// match decision types together with their params
type sameDecision struct{ basicDecision }
type replaceDecision struct{ basicDecision }
type matchDecision struct{ basicDecision }
type permuteDecision struct {
	basicDecision
	permutation []int
}

//min is a convinience method for choosing the cheapest decision
func min(self, candidate decisionType) decisionType {
	if self.cost() > candidate.cost() {
		return candidate
	} else {
		return self
	}
}

//type for cache
type keyType struct{ k1, k2 ID }

//cache for diff computation
type cacheStorage struct {
	decisions map[keyType]decisionType
	counts     map[ID]int
	changes   Changelist
}

func makeCacheStorage() *cacheStorage {
	return &cacheStorage{
		decisions: make(map[keyType]decisionType),
		counts: make(map[ID]int),
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
func (ds *cacheStorage) decideAction(src, dst nodes.Node) decisionType {
	label := keyType{nodes.UniqueKey(src), nodes.UniqueKey(dst)}

	if val, ok := ds.decisions[label]; ok {
		return val
	}

	// the default (but not always optimal) decision is to ignore the existance of src and just
	// create the desired dst; implemented below
	cost := ds.nodeSize(dst) + 1

	var bestDecision decisionType
	bestDecision = replaceDecision{basicDecision{cost}}

	if nodes.KindOf(src) != nodes.KindOf(dst) {
		ds.decisions[label] = bestDecision
		return bestDecision
	}

	// from now on src.(type) == dst.(type)

	switch src := src.(type) {

	case nodes.Value:
		dst := dst.(nodes.Value)
		cost = 0
		if src == dst {
			bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})
		} else {
			cost = 1
			bestDecision = min(bestDecision, replaceDecision{basicDecision{cost}})
		}

	case nodes.Object:
		dst := dst.(nodes.Object)
		cost = 0

		// the code below iterates over each of the keys from src and dst exactly once, that's why
		// keys isn't reset between iterations.
		keys := make(map[string]bool)
		iterate := func(keyset nodes.Object) {
			for key := range keyset {
				if in := keys[key]; !in {
					keys[key] = true
					cost += ds.decideAction(src[key], dst[key]).cost()
				}
			}
		}
		iterate(src)
		iterate(dst)

		if cost == 0 {
			bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})
		} else {
			bestDecision = min(bestDecision, matchDecision{basicDecision{cost}})
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
				bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})
				break
			}
		}

		if len(src) != len(dst) {
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

		bestDecision = min(bestDecision, permuteDecision{basicDecision{cost}, res.InRow})

	case nil:
		cost = 0
		bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})

	default:
		panic(fmt.Errorf("unknown node type %T", src))
	}

	ds.decisions[label] = bestDecision
	return ds.decisions[label]
}

func (ds *cacheStorage) push(change Change) {
	ds.changes = append(ds.changes, change)
}

func (ds *cacheStorage) createRec(node nodes.Node) {
	switch n := node.(type) {
	case nodes.Object:
		for _, v := range n {
			ds.createRec(v)
		}

	case nodes.Array:
		for _, v := range n {
			ds.createRec(v)
		}

	default:
		// values and nils are not saved separately
		return
	}
	ds.push(Create{Node: node})
}

func (ds *cacheStorage) generateDifference(src, dst nodes.Node, parentID ID, parentKey Key) {
	switch d := ds.decideAction(src, dst).(type) {

	// no action required if same
	case sameDecision:

	case replaceDecision:
		// remove src (no action?) and create dst + attach it
		if dst != nil {
			ds.createRec(dst)
			ds.push(Attach{Parent: parentID, Key: parentKey, Child: nodes.UniqueKey(dst)})
		} else {
			ds.push(Attach{Parent: parentID, Key: parentKey, Child: nil})
		}

	case matchDecision:
		src, dst := src.(nodes.Object), dst.(nodes.Object)
		keys := make(map[string]bool)
		iterate := func(keyset nodes.Object) {
			for key := range keyset {
				if in := keys[key]; !in {
					keys[key] = true
					if _, ok := dst[key]; !ok {
						ds.push(Deatach{Parent: nodes.UniqueKey(src), Key: String(key)})
					} else  if _, ok := src[key]; !ok {
						ds.createRec(dst[key])
						ds.push(Attach{
							Parent: nodes.UniqueKey(src),
							Key: String(key),
							Child: nodes.UniqueKey(dst[key]),
						})
					} else {
						ds.generateDifference(
							src[key], dst[key], nodes.UniqueKey(src), String(key))
					}
				}
			}
		}
		iterate(src)
		iterate(dst)

	case permuteDecision:
		src, dst := src.(nodes.Array), dst.(nodes.Array)
		l := len(dst) - len(src)
		// add possible nils to src
		if l > 0 {
			arr := make(nodes.Array, len(dst))
			copy(arr, src)
			src = arr
		}
		recreate := false
		if l != 0 {
			recreate = true
		}
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
			ds.push(Create{Node: src}) // TODO: not create, only mutate...
			ds.push(Attach{Parent: parentID, Key: parentKey, Child: nodes.UniqueKey(src)})
		}
		for i := 0; i < len(dst); i++ {
			ds.generateDifference(src[i], dst[i], nodes.UniqueKey(src), Int(i))
		}

	default:
		panic(fmt.Errorf("unknown decision %v", d))
	}
}

func Cost(src, dst nodes.Node) int {
	ds := makeCacheStorage()
	return ds.decideAction(src, dst).cost()
}

func Changes(src, dst nodes.Node) Changelist {
	ds := makeCacheStorage()
	ds.generateDifference(src, dst, nil, Int(0))
	return ds.changes
}
