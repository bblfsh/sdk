package diff

import (
	// TODO: integrate with https://github.com/src-d/lapjv if perf is unacceptable
	"fmt"
	"github.com/heetch/lapjv"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type decisionType interface {
	isDecisionType()
}

type basicDecisionType struct{}

func (_ basicDecisionType) isDecisionType() {}

type decision struct {
	cost     int
	decision decisionType
}

// match decision types together with their params
type sameDecision struct{ basicDecisionType }
type replaceDecision struct{ basicDecisionType }
type matchDecision struct{ basicDecisionType }
type permuteDecision struct {
	basicDecisionType
	permutation []int
}

//convinience method for choosing the cheapest decision
func (self *decision) minEq(candidate *decision) {
	if _, ok := candidate.decision.(replaceDecision); ok || self.cost > candidate.cost {
		self.cost, self.decision = candidate.cost, candidate.decision
	}
}

//type for cache
type keyType struct{ k1, k2 ID }

//cache for diff computation
type cacheStorage struct {
	decisions map[keyType]decision
	sizes     map[ID]int
	changes   Changelist
}

func emptyCacheStorage() cacheStorage {
	return cacheStorage{
		make(map[keyType]decision),
		make(map[ID]int),
		make([]Change, 0),
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
func (ds *cacheStorage) decideAction(src, dst nodes.Node) decision {
	label := keyType{nodes.UniqueKey(src), nodes.UniqueKey(dst)}

	if val, ok := ds.decisions[label]; ok {
		return val
	}

	// one can always just create the dst ignoring the src
	cost := ds.nodeSize(dst) + 1
	bestDecision := decision{cost, replaceDecision{}}

	if nodes.KindOf(src) != nodes.KindOf(dst) {
		ds.decisions[label] = bestDecision
		return bestDecision
	}

	switch src.(type) {

	case nodes.Value:
		src, dst := src.(nodes.Value), dst.(nodes.Value)
		if src == dst {
			bestDecision.minEq(&decision{0, sameDecision{}})
		} else {
			bestDecision.minEq(&decision{1, replaceDecision{}})
		}

	case nodes.Object:
		src, dst := src.(nodes.Object), dst.(nodes.Object)
		cost = 0

		keys := make(map[string]bool)
		iterate := func(keyset nodes.Object) {
			for key := range keyset {
				if in := keys[key]; !in {
					keys[key] = true
					cost += ds.decideAction(src[key], dst[key]).cost
				}
			}
		}
		iterate(src)
		iterate(dst)

		if cost == 0 {
			bestDecision.minEq(&decision{cost, sameDecision{}})
		} else {
			bestDecision.minEq(&decision{cost, matchDecision{}})
		}

	case nodes.Array:
		src, dst := src.(nodes.Array), dst.(nodes.Array)
		cost = 0
		if len(src) == len(dst) && func() int {
			sum := 0
			for i := range src {
				sum += ds.decideAction(src[i], dst[i]).cost
			}
			return sum
		}() == 0 {
			bestDecision.minEq(&decision{cost, sameDecision{}})
			break
		}

		if len(src) != len(dst) {
			cost = 2
		}

		if len(src) < len(dst) {
			l := len(dst) - len(src)
			for i := 0; i < l; i++ {
				src = append(src, nil)
			}
		} else {
			l := len(src) - len(dst)
			for i := 0; i < l; i++ {
				dst = append(dst, nil)
			}
		}
		n := len(src)
		m := make([][]int, n)
		for i := 0; i < n; i++ {
			m[i] = make([]int, n)
			for j := 0; j < n; j++ {
				m[i][j] = ds.decideAction(src[i], dst[j]).cost
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

		bestDecision.minEq(&decision{cost, permuteDecision{permutation: res.InRow}})

	case nil:
		bestDecision.minEq(&decision{0, sameDecision{}})

	default:
		panic(fmt.Sprintf("unknown node type %T", src))
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
	switch d := decision.decision.(type) {

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
			for i := 0; i < l; i++ {
				src = append(src, nil)
			}
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
			newsrc := make([]nodes.Node, 0, len(dst))
			for i := 0; i < len(dst); i++ {
				newsrc = append(newsrc, src[d.permutation[i]])
			}
			src = newsrc
			ds.push(&Create{node: src}) // TODO: not create, only mutate...
			ds.push(&Attach{parent: parentID, key: parentKey, child: nodes.UniqueKey(src)})
		}
		for i := 0; i < len(dst); i++ {
			ds.generateDifference(src[i], dst[i], nodes.UniqueKey(src), Int(i))
		}

	default:
		panic(fmt.Sprintf("unknown decision %v", d))
	}
}

func Cost(src, dst nodes.Node) int {
	ds := emptyCacheStorage()
	return ds.decideAction(src, dst).cost
}

func Changes(src, dst nodes.Node) Changelist {
	ds := emptyCacheStorage()
	ds.generateDifference(src, dst, nil, Int(0))
	return ds.changes
}
