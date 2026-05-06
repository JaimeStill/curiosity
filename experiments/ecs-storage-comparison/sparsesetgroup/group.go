package sparsesetgroup

import "ecs-storage-comparison/storage"

type group struct {
	set     []storage.ComponentID
	columns []*column
	size    int
}

func (g *group) coveredBy(components []storage.ComponentValue) bool {
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

func (g *group) matches(set []storage.ComponentID) bool {
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
