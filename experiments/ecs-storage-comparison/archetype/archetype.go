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

func (c *column) appendValue(src unsafe.Pointer) {
	c.data = append(c.data, unsafe.Slice((*byte)(src), c.size)...)
}

type archetype struct {
	signature storage.Signature
	cids      []storage.ComponentID
	columns   []column
	entities  []storage.EntityID
}

func (a *archetype) columnFor(cid storage.ComponentID) *column {
	for i := range a.cids {
		if a.cids[i] == cid {
			return &a.columns[i]
		}
	}
	return nil
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

var _ storage.Storage = (*Storage)(nil)

func New() *Storage {
	return &Storage{
		archetypes: make(map[storage.Signature]*archetype),
		locations:  make(map[storage.EntityID]location),
	}
}

func (s *Storage) ApplyDeferred() {
	panic("not implemented")
}

func (s *Storage) Attach(id storage.EntityID, cid storage.ComponentID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) Detach(id storage.EntityID, cid storage.ComponentID) {
	panic("not implemented")
}

func (s *Storage) Despawn(id storage.EntityID) {
	panic("not implemented")
}

func (s *Storage) Spawn(components []storage.ComponentValue) storage.EntityID {
	var sig storage.Signature
	cids := make([]storage.ComponentID, len(components))
	for i, cv := range components {
		sig.Set(cv.ID)
		cids[i] = cv.ID
	}
	arch := s.getOrCreateArchetype(sig, cids)

	s.nextID++
	id := s.nextID
	row := len(arch.entities)
	arch.entities = append(arch.entities, id)

	for _, cv := range components {
		arch.
			columnFor(cv.ID).
			appendValue(cv.Data)
	}

	s.locations[id] = location{
		arch: arch,
		row:  row,
	}
	return id
}

func (s *Storage) Query(set []storage.ComponentID) storage.Iterator {
	sig := storage.SignatureOf(set)
	matches := make([]*archetype, 0, len(s.archetypes))
	for _, a := range s.archetypes {
		if a.signature.Contains(sig) {
			matches = append(matches, a)
		}
	}
	return &iterator{
		matches: matches,
		index:   0,
		row:     -1,
	}
}

func (s *Storage) Read(id storage.EntityID, cid storage.ComponentID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id storage.EntityID, cid storage.ComponentID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) getOrCreateArchetype(sig storage.Signature, cids []storage.ComponentID) *archetype {
	if a, ok := s.archetypes[sig]; ok {
		return a
	}
	slices.Sort(cids)

	cols := make([]column, len(cids))
	for i, cid := range cids {
		cols[i] = column{
			cid:  cid,
			size: storage.ComponentType(cid).Size(),
			data: nil,
		}
	}

	a := &archetype{
		signature: sig,
		cids:      cids,
		columns:   cols,
	}
	s.archetypes[sig] = a
	return a
}
