package sparsesetgroup

import (
	"slices"
	"unsafe"

	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type column struct {
	cid      component.ID
	size     uintptr
	sparse   []int32
	dense    []byte
	entities []entity.ID
}

func (c *column) appendValue(src unsafe.Pointer) {
	c.dense = append(c.dense, unsafe.Slice((*byte)(src), c.size)...)
}

func (c *column) ensureSparseCapacity(id entity.ID) {
	for len(c.sparse) <= int(id) {
		c.sparse = append(c.sparse, -1)
	}
}

func (c *column) swapDenseRows(i, j int32) {
	if i == j {
		return
	}
	a := uintptr(i) * c.size
	b := uintptr(j) * c.size
	da := c.dense[a : a+c.size]
	db := c.dense[b : b+c.size]
	for k := uintptr(0); k < c.size; k++ {
		da[k], db[k] = db[k], da[k]
	}
}

func (c *column) swapRemove(id entity.ID) {
	r := int(c.sparse[id])
	lastRow := len(c.entities) - 1
	if r != lastRow {
		movedID := c.entities[lastRow]
		c.entities[r] = movedID
		dst := uintptr(r) * c.size
		src := uintptr(lastRow) * c.size
		copy(c.dense[dst:dst+c.size], c.dense[src:src+c.size])
		c.sparse[movedID] = int32(r)
	}
	c.entities = c.entities[:lastRow]
	c.dense = c.dense[:uintptr(lastRow)*c.size]
	c.sparse[id] = -1
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
	columns   map[component.ID]*column
	groups    []*group
	alloc     *entity.Allocator
	despawned []entity.ID
	attached  []attachComponent
	detached  []detachComponent
}

var _ storage.Storage = (*Storage)(nil)

func New(alloc *entity.Allocator, groups [][]component.ID) *Storage {
	s := &Storage{
		columns: make(map[component.ID]*column),
		alloc:   alloc,
	}
	for _, set := range groups {
		g := &group{
			set:     set,
			columns: make([]*column, len(set)),
		}
		for i, cid := range set {
			g.columns[i] = s.getOrCreateColumn(cid)
		}
		s.groups = append(s.groups, g)
	}
	return s
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
	id := s.alloc.Allocate()
	for _, cv := range components {
		c := s.getOrCreateColumn(cv.ID)
		c.ensureSparseCapacity(id)
		c.sparse[id] = int32(len(c.entities))
		c.entities = append(c.entities, id)
		c.appendValue(cv.Data)
	}
	for _, g := range s.groups {
		if !g.coveredBy(components) {
			continue
		}
		for _, c := range g.columns {
			row := int32(g.size)
			tail := int32(len(c.entities) - 1)
			if row == tail {
				continue
			}
			other := c.entities[row]
			c.entities[row], c.entities[tail] = c.entities[tail], c.entities[row]
			c.swapDenseRows(row, tail)
			c.sparse[id] = row
			c.sparse[other] = tail
		}
		g.size++
	}
	return id
}

func (s *Storage) Query(set []component.ID) storage.Iterator {
	if len(set) == 0 {
		return &iterator{}
	}
	refs := make([]queryRef, len(set))
	for i, cid := range set {
		c, ok := s.columns[cid]
		if !ok {
			return &iterator{}
		}
		refs[i].col = c
	}
	for _, g := range s.groups {
		if g.matches(set) {
			return &iterator{
				refs:  refs,
				row:   -1,
				group: g,
			}
		}
	}
	driver := 0
	for i := 1; i < len(refs); i++ {
		if len(refs[i].col.entities) < len(refs[driver].col.entities) {
			driver = i
		}
	}
	return &iterator{
		refs:   refs,
		driver: driver,
		row:    -1,
	}
}

func (s *Storage) Read(id entity.ID, cid component.ID) (unsafe.Pointer, bool) {
	c, ok := s.columns[cid]
	if !ok {
		return nil, false
	}
	if int(id) >= len(c.sparse) || c.sparse[id] == -1 {
		return nil, false
	}
	row := int(c.sparse[id])
	return unsafe.Pointer(&c.dense[uintptr(row)*c.size]), true
}

func (s *Storage) Write(id entity.ID, cid component.ID, data unsafe.Pointer) {
	c, ok := s.columns[cid]
	if !ok {
		return
	}
	if int(id) >= len(c.sparse) || c.sparse[id] == -1 {
		return
	}
	row := int(c.sparse[id])
	dst := uintptr(row) * c.size
	copy(c.dense[dst:dst+c.size], unsafe.Slice((*byte)(data), c.size))
}

func (s *Storage) applyDespawn(id entity.ID) {
	for _, g := range s.groups {
		c0 := g.columns[0]
		if int(id) >= len(c0.sparse) {
			continue
		}
		r := c0.sparse[id]
		if r < 0 || r >= int32(g.size) {
			continue
		}
		boundary := int32(g.size - 1)
		if r != boundary {
			other := c0.entities[boundary]
			for _, c := range g.columns {
				c.entities[r], c.entities[boundary] = c.entities[boundary], c.entities[r]
				c.swapDenseRows(r, boundary)
				c.sparse[id] = boundary
				c.sparse[other] = r
			}
		}
		g.size--
	}
	for _, c := range s.columns {
		if int(id) < len(c.sparse) && c.sparse[id] != -1 {
			c.swapRemove(id)
		}
	}
	s.alloc.Free(id)
}

func (s *Storage) applyAttach(op attachComponent) {
	c := s.getOrCreateColumn(op.cid)
	c.ensureSparseCapacity(op.id)
	if c.sparse[op.id] != -1 {
		row := int(c.sparse[op.id])
		dst := uintptr(row) * c.size
		copy(c.dense[dst:dst+c.size], op.data)
		return
	}
	c.sparse[op.id] = int32(len(c.entities))
	c.entities = append(c.entities, op.id)
	c.dense = append(c.dense, op.data...)
	s.enterPrefixIfCovered(op.id, op.cid)
}

func (s *Storage) applyDetach(op detachComponent) {
	c, ok := s.columns[op.cid]
	if !ok {
		return
	}
	if int(op.id) >= len(c.sparse) || c.sparse[op.id] == -1 {
		return
	}
	s.exitPrefixIfInPrefix(op.id, op.cid)
	c.swapRemove(op.id)
}

func (s *Storage) enterPrefixIfCovered(id entity.ID, cid component.ID) {
	for _, g := range s.groups {
		cidInGroup := slices.Contains(g.set, cid)
		if !cidInGroup {
			continue
		}
		fullyCovered := true
		for _, c := range g.columns {
			if int(id) >= len(c.sparse) || c.sparse[id] == -1 {
				fullyCovered = false
				break
			}
		}
		if !fullyCovered {
			continue
		}
		boundary := int32(g.size)
		for _, c := range g.columns {
			r := c.sparse[id]
			if r == boundary {
				continue
			}
			otherID := c.entities[boundary]
			c.entities[r], c.entities[boundary] = c.entities[boundary], c.entities[r]
			c.swapDenseRows(r, boundary)
			c.sparse[id] = boundary
			c.sparse[otherID] = r
		}
		g.size++
	}
}

func (s *Storage) exitPrefixIfInPrefix(id entity.ID, cid component.ID) {
	for _, g := range s.groups {
		if !slices.Contains(g.set, cid) {
			continue
		}
		c0 := g.columns[0]
		if int(id) >= len(c0.sparse) {
			continue
		}
		r := c0.sparse[id]
		if r < 0 || r >= int32(g.size) {
			continue
		}
		boundary := int32(g.size - 1)
		if r != boundary {
			otherID := c0.entities[boundary]
			for _, c := range g.columns {
				c.entities[r], c.entities[boundary] = c.entities[boundary], c.entities[r]
				c.swapDenseRows(r, boundary)
				c.sparse[id] = boundary
				c.sparse[otherID] = r
			}
		}
		g.size--
	}
}

func (s *Storage) getOrCreateColumn(cid component.ID) *column {
	if c, ok := s.columns[cid]; ok {
		return c
	}
	c := &column{
		cid:  cid,
		size: component.TypeOf(cid).Size(),
	}
	s.columns[cid] = c
	return c
}
