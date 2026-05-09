package sparsesetgroup

import "ecs-storage-comparison/component"

type group struct {
	set     []component.ID
	columns []*column
	size    int
}

func (g *group) coveredBy(components []component.Value) bool {
	for _, cid := range g.set {
		found := false
		for _, cv := range components {
			if cv.ID == cid {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (g *group) matches(set []component.ID) bool {
	if len(g.set) != len(set) {
		return false
	}
	for _, cid := range g.set {
		found := false
		for _, q := range set {
			if q == cid {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
