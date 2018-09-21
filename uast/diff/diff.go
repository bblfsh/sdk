package diff

import (
	"fmt"
	"github.com/heetch/lapjv"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type decisionType interface {
	isDecisionType()
	cost() int
}

type basicDecision struct {
	privateCost int
}

func (basicDecision) isDecisionType() {}

func (b basicDecision) cost() int { return b.privateCost }

// match decision types together with their params
type sameDecision struct{ basicDecision }
type replaceDecision struct{ basicDecision }
type matchDecision struct{ basicDecision }
type permuteDecision struct {
	basicDecision
	permutation []int
}

//convinience method for choosing the cheapest decision
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
	sizes     map[ID]int
	changes   Changelist
}

func emptyCacheStorage() cacheStorage {
	return cacheStorage{
		make(map[keyType]decisionType),
		make(map[ID]int),
		nil,
	}
}

// caches (for perf) tree size (node count)
func (ds *cacheStorage) nodeSize(n nodes.Node) int {
	label := nodes.UniqueKey(n)
	if cnt, ok := ds.sizes[label]; ok {
		return cnt
	}
	ret := nodes.Count(n, nodes.KindsNotNil)
	ds.sizes[label] = ret
	return ret
}

// find the cheapest way to naively match src and dst and return the action with its combined cost
func (ds *cacheStorage) decideAction(src, dst nodes.Node) decisionType {
	label := keyType{nodes.UniqueKey(src), nodes.UniqueKey(dst)}

	if val, ok := ds.decisions[label]; ok {
		return val
	}

	// one can always just create the dst ignoring the src
	cost := ds.nodeSize(dst) + 1

	var bestDecision decisionType
	bestDecision = replaceDecision{basicDecision{cost}}

	if nodes.KindOf(src) != nodes.KindOf(dst) {
		ds.decisions[label] = bestDecision
		return bestDecision
	}

	switch src.(type) {

	case nodes.Value:
		src, dst := src.(nodes.Value), dst.(nodes.Value)
		cost = 0
		if src == dst {
			bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})
		} else {
			cost = 1
			bestDecision = min(bestDecision, replaceDecision{basicDecision{cost}})
		}

	case nodes.Object:
		src, dst := src.(nodes.Object), dst.(nodes.Object)
		cost = 0

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
		src, dst := src.(nodes.Array), dst.(nodes.Array)
		cost = 0
		if len(src) == len(dst) && func() int {
			sum := 0
			for i := range src {
				sum += ds.decideAction(src[i], dst[i]).cost()
			}
			return sum
		}() == 0 {
			bestDecision = min(bestDecision, sameDecision{basicDecision{cost}})
			break
		}

		if len(src) != len(dst) {
			cost = 2
		}

		if len(src) < len(dst) {
			arr := make(nodes.Array, len(dst))
			copy(arr, src)
			src = arr
		} else {
			arr := make(nodes.Array, len(src))
			copy(arr, dst)
			dst = arr
		}
		n := len(src)
		m := make([][]int, n)
		for i := 0; i < n; i++ {
			m[i] = make([]int, n)
			for j := 0; j < n; j++ {
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
		//values and nils are not saved separately
		return
	}
	ds.push(&Create{node: node})
}

func (ds *cacheStorage) generateDifference(src, dst nodes.Node, parentID ID, parentKey Key) {
	decision := ds.decideAction(src, dst)
	switch d := decision.(type) {

	//no action required if same
	case sameDecision:

	case replaceDecision:
		//remove src (no action?) and create dst + attach it
		if dst != nil {
			ds.createRec(dst)
			ds.push(&Attach{parent: parentID, key: parentKey, child: nodes.UniqueKey(dst)})
		} else {
			ds.push(&Attach{parent: parentID, key: parentKey, child: nil})
		}

	case matchDecision:
		src, dst := src.(nodes.Object), dst.(nodes.Object)
		keys := make(map[string]bool)
		iterate := func(keyset nodes.Object) {
			for key := range keyset {
				if in := keys[key]; !in {
					keys[key] = true
					if _, ok := dst[key]; !ok {
						ds.push(&Deattach{parent: nodes.UniqueKey(src), key: String(key)})
					} else {
						ds.generateDifference(src[key], dst[key], nodes.UniqueKey(src), String(key))
					}
				}
			}
		}
		iterate(src)
		iterate(dst)

	case permuteDecision:
		src, dst := src.(nodes.Array), dst.(nodes.Array)
		l := len(dst) - len(src)
		//add possible nils to src
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
			//recreate src with right perm
			newsrc := make([]nodes.Node, len(dst))
			for i := 0; i < len(dst); i++ {
				newsrc[i] = src[d.permutation[i]]
			}
			src = newsrc
			ds.push(&Create{node: src}) // TODO: not create, only mutate...
			ds.push(&Attach{parent: parentID, key: parentKey, child: nodes.UniqueKey(src)})
		}
		for i := 0; i < len(dst); i++ {
			ds.generateDifference(src[i], dst[i], nodes.UniqueKey(src), Int(i))
		}

	default:
		panic(fmt.Errorf("unknown decision %v", d))
	}
}

func Cost(src, dst nodes.Node) int {
	ds := emptyCacheStorage()
	return ds.decideAction(src, dst).cost()
}

func Changes(src, dst nodes.Node) Changelist {
	ds := emptyCacheStorage()
	ds.generateDifference(src, dst, nil, Int(0))
	return ds.changes
}
