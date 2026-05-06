package sparsesetslice

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

type Storage struct {
	columns map[storage.ComponentID]*column
	nextID  storage.EntityID
}

var _ storage.Storage = (*Storage)(nil)

func New() *Storage {
	return &Storage{
		columns: make(map[storage.ComponentID]*column),
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
	s.nextID++
	id := s.nextID
	for _, cv := range components {
		c := s.getOrCreateColumn(cv.ID)
		c.ensureSparseCapacity(id)
		c.sparse[id] = int32(len(c.entities))
		c.entities = append(c.entities, id)
		c.appendValue(cv.Data)
	}
	return id
}

func (s *Storage) Read(id storage.EntityID, cid storage.ComponentID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id storage.EntityID, cid storage.ComponentID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) Query(set []storage.ComponentID) storage.Iterator {
	refs := make([]queryRef, len(set))
	driver := 0
	for i, cid := range set {
		c, ok := s.columns[cid]
		if !ok {
			return &iterator{}
		}
		refs[i].col = c
		if len(c.entities) < len(refs[driver].col.entities) {
			driver = i
		}
	}
	return &iterator{
		refs:   refs,
		driver: driver,
		row:    -1,
	}
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
