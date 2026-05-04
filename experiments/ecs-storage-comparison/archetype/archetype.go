package archetype

import (
	"slices"
	"unsafe"

	"ecs-storage-comparison/storage"
)

type column struct {
	cid  storage.ComponentID
	size uintptr
	data []byte
}

type archetype struct {
	signature storage.Signature
	cids      []storage.ComponentID
	columns   []column
	entities  []storage.EntityID
}

type location struct {
	arch *archetype
	row  int
}

type Storage struct {
	archetypes map[storage.Signature]*archetype
	locations  map[storage.EntityID]location
	nextID     storage.EntityID
}

func New() *Storage {
	return &Storage{
		archetypes: make(map[storage.Signature]*archetype),
		locations:  make(map[storage.EntityID]location),
	}
}

func (c *column) appendValue(src unsafe.Pointer) {
	c.data = append(c.data, unsafe.Slice((*byte)(src), c.size)...)
}

func (a *archetype) columnFor(cid storage.ComponentID) *column {
	for i := range a.cids {
		if a.cids[i] == cid {
			return &a.columns[i]
		}
	}
	return nil
}

func (s *Storage) getOrCreateArchetype(sig storage.Signature, cids []storage.ComponentID) *archetype {
	if a, ok := s.archetypes[sig]; ok {
		return a
	}
	sortedCIDs := slices.Clone(cids)
	slices.Sort(sortedCIDs)

	cols := make([]column, len(sortedCIDs))
	for i, cid := range sortedCIDs {
		cols[i] = column{
			cid:  cid,
			size: storage.ComponentType(cid).Size(),
			data: nil,
		}
	}

	a := &archetype{
		signature: sig,
		cids:      sortedCIDs,
		columns:   cols,
	}
	s.archetypes[sig] = a
	return a
}
