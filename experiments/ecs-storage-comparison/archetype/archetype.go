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

type attachComponent struct {
	id   entity.ID
	cid  component.ID
	data []byte
}

type detachComponent struct {
	id  entity.ID
	cid component.ID
}

type Storage struct {
	archetypes map[component.Signature]*archetype
	locations  []location
	alloc      *entity.Allocator
	despawned  []entity.ID
	attached   []attachComponent
	detached   []detachComponent
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
	for _, op := range s.detached {
		s.applyDetach(op)
	}
	for _, op := range s.attached {
		s.applyAttach(op)
	}
	s.despawned = s.despawned[:0]
	s.detached = s.detached[:0]
	s.attached = s.attached[:0]
}

func (s *Storage) Attach(id entity.ID, cid component.ID, data unsafe.Pointer) {
	size := component.TypeOf(cid).Size()
	buf := make([]byte, size)
	copy(buf, unsafe.Slice((*byte)(data), size))
	s.attached = append(s.attached, attachComponent{id: id, cid: cid, data: buf})
}

func (s *Storage) Detach(id entity.ID, cid component.ID) {
	s.detached = append(s.detached, detachComponent{id: id, cid: cid})
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
	if int(id) >= len(s.locations) {
		return nil, false
	}
	loc := s.locations[id]
	if loc.arch == nil {
		return nil, false
	}
	col := loc.arch.columnFor(cid)
	if col == nil {
		return nil, false
	}
	return unsafe.Pointer(&col.data[uintptr(loc.row)*col.size]), true
}

func (s *Storage) Write(id entity.ID, cid component.ID, data unsafe.Pointer) {
	if int(id) >= len(s.locations) {
		return
	}
	loc := s.locations[id]
	if loc.arch == nil {
		return
	}
	col := loc.arch.columnFor(cid)
	if col == nil {
		return
	}
	dst := uintptr(loc.row) * col.size
	copy(col.data[dst:dst+col.size], unsafe.Slice((*byte)(data), col.size))
}

func (s *Storage) applyDespawn(id entity.ID) {
	loc := s.locations[id]
	s.swapRemove(loc.arch, loc.row)
	s.locations[id] = location{}
	s.alloc.Free(id)
}

func (s *Storage) applyAttach(op attachComponent) {
	loc := s.locations[op.id]
	oldArch := loc.arch

	if existing := oldArch.columnFor(op.cid); existing != nil {
		dst := uintptr(loc.row) * existing.size
		copy(existing.data[dst:dst+existing.size], op.data)
		return
	}

	newSig := oldArch.signature
	newSig.Set(op.cid)
	newCids := make([]component.ID, len(oldArch.cids)+1)
	copy(newCids, oldArch.cids)
	newCids[len(oldArch.cids)] = op.cid

	newArch := s.getOrCreateArchetype(newSig, newCids)
	newRow := len(newArch.entities)
	newArch.entities = append(newArch.entities, op.id)

	for i := range oldArch.columns {
		oldCol := &oldArch.columns[i]
		newCol := newArch.columnFor(oldCol.cid)
		oldOff := uintptr(loc.row) * oldCol.size
		newCol.data = append(newCol.data, oldCol.data[oldOff:oldOff+oldCol.size]...)
	}

	newCol := newArch.columnFor(op.cid)
	newCol.data = append(newCol.data, op.data...)

	s.swapRemove(oldArch, loc.row)
	s.locations[op.id] = location{arch: newArch, row: newRow}
}

func (s *Storage) applyDetach(op detachComponent) {
	loc := s.locations[op.id]
	oldArch := loc.arch

	if oldArch.columnFor(op.cid) == nil {
		return
	}

	var newSig component.Signature
	newCids := make([]component.ID, 0, len(oldArch.cids)-1)
	for _, cid := range oldArch.cids {
		if cid == op.cid {
			continue
		}
		newSig.Set(cid)
		newCids = append(newCids, cid)
	}

	newArch := s.getOrCreateArchetype(newSig, newCids)
	newRow := len(newArch.entities)
	newArch.entities = append(newArch.entities, op.id)

	for i := range oldArch.columns {
		oldCol := &oldArch.columns[i]
		if oldCol.cid == op.cid {
			continue
		}
		newCol := newArch.columnFor(oldCol.cid)
		oldOff := uintptr(loc.row) * oldCol.size
		newCol.data = append(newCol.data, oldCol.data[oldOff:oldOff+oldCol.size]...)
	}

	s.swapRemove(oldArch, loc.row)
	s.locations[op.id] = location{arch: newArch, row: newRow}
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

func (s *Storage) swapRemove(arch *archetype, row int) {
	lastRow := len(arch.entities) - 1
	if row != lastRow {
		movedID := arch.entities[lastRow]
		arch.entities[row] = movedID
		for i := range arch.columns {
			col := &arch.columns[i]
			dst := uintptr(row) * col.size
			src := uintptr(lastRow) * col.size
			copy(col.data[dst:dst+col.size], col.data[src:src+col.size])
		}
		s.locations[movedID] = location{arch: arch, row: row}
	}
	arch.entities = arch.entities[:lastRow]
	for i := range arch.columns {
		col := &arch.columns[i]
		col.data = col.data[:uintptr(lastRow)*col.size]
	}
}
