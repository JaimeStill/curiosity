package sparsesetgroup

import (
	"unsafe"

	"ecs-storage-comparison/storage"
)

type column struct {
	cid      storage.ComponentID
	size     uintptr
	sparse   []int32
	dense    []byte
	entities []storage.EntityID
}

func (c *column) appendValue(src unsafe.Pointer) {
	c.dense = append(c.dense, unsafe.Slice((*byte)(src), c.size)...)
}

func (c *column) ensureSparseCapacity(id storage.EntityID) {
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

type Storage struct {
	columns map[storage.ComponentID]*column
	groups  []*group
	nextID  storage.EntityID
}

var _ storage.Storage = (*Storage)(nil)

func New(groups [][]storage.ComponentID) *Storage {
	s := &Storage{
		columns: make(map[storage.ComponentID]*column),
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
	s.nextID++
	id := s.nextID
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

func (s *Storage) Query(set []storage.ComponentID) storage.Iterator {
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

func (s *Storage) Read(id storage.EntityID, cid storage.ComponentID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id storage.EntityID, cid storage.ComponentID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) getOrCreateColumn(cid storage.ComponentID) *column {
	if c, ok := s.columns[cid]; ok {
		return c
	}
	c := &column{
		cid:  cid,
		size: storage.ComponentType(cid).Size(),
	}
	s.columns[cid] = c
	return c
}
