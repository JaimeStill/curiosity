package archetype

import (
	"slices"
	"unsafe"

	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type column struct {
	cid  component.ID
	size uintptr
	data []byte
}

func (c *column) appendValue(src unsafe.Pointer) {
	c.data = append(c.data, unsafe.Slice((*byte)(src), c.size)...)
}

type archetype struct {
	signature component.Signature
	cids      []component.ID
	columns   []column
	entities  []entity.ID
}

func (a *archetype) columnFor(cid component.ID) *column {
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
	archetypes map[component.Signature]*archetype
	locations  []location
	alloc      *entity.Allocator
	despawned  []entity.ID
}

var _ storage.Storage = (*Storage)(nil)

func New(alloc *entity.Allocator) *Storage {
	return &Storage{
		archetypes: make(map[component.Signature]*archetype),
		alloc:      alloc,
	}
}

func (s *Storage) ApplyDeferred() {
	for _, id := range s.despawned {
		s.applyDespawn(id)
	}
	s.despawned = s.despawned[:0]
}

func (s *Storage) Attach(id entity.ID, cid component.ID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) Detach(id entity.ID, cid component.ID) {
	panic("not implemented")
}

func (s *Storage) Despawn(id entity.ID) {
	s.despawned = append(s.despawned, id)
}

func (s *Storage) Spawn(components []component.Value) entity.ID {
	var sig component.Signature
	cids := make([]component.ID, len(components))
	for i, cv := range components {
		sig.Set(cv.ID)
		cids[i] = cv.ID
	}
	arch := s.getOrCreateArchetype(sig, cids)

	id := s.alloc.Allocate()
	row := len(arch.entities)
	arch.entities = append(arch.entities, id)

	for _, cv := range components {
		arch.
			columnFor(cv.ID).
			appendValue(cv.Data)
	}

	if grow := int(id) + 1; grow > len(s.locations) {
		s.locations = slices.Grow(s.locations, grow-len(s.locations))[:grow]
	}
	s.locations[id] = location{
		arch: arch,
		row:  row,
	}
	return id
}

func (s *Storage) Query(set []component.ID) storage.Iterator {
	sig := component.SignatureOf(set)
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

func (s *Storage) Read(id entity.ID, cid component.ID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id entity.ID, cid component.ID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) applyDespawn(id entity.ID) {
	loc := s.locations[id]
	arch := loc.arch
	lastRow := len(arch.entities) - 1

	if loc.row != lastRow {
		movedID := arch.entities[lastRow]
		arch.entities[loc.row] = movedID
		for i := range arch.columns {
			col := &arch.columns[i]
			dst := uintptr(loc.row) * col.size
			src := uintptr(lastRow) * col.size
			copy(col.data[dst:dst+col.size], col.data[src:src+col.size])
		}
		s.locations[movedID] = location{arch: arch, row: loc.row}
	}

	arch.entities = arch.entities[:lastRow]
	for i := range arch.columns {
		col := &arch.columns[i]
		col.data = col.data[:uintptr(lastRow)*col.size]
	}

	s.locations[id] = location{}
	s.alloc.Free(id)
}

func (s *Storage) getOrCreateArchetype(sig component.Signature, cids []component.ID) *archetype {
	if a, ok := s.archetypes[sig]; ok {
		return a
	}
	slices.Sort(cids)

	cols := make([]column, len(cids))
	for i, cid := range cids {
		cols[i] = column{
			cid:  cid,
			size: component.TypeOf(cid).Size(),
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
