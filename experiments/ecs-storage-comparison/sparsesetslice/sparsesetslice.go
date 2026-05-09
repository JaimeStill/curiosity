package sparsesetslice

import (
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

type Storage struct {
	columns   map[component.ID]*column
	alloc     *entity.Allocator
	despawned []entity.ID
}

var _ storage.Storage = (*Storage)(nil)

func New(alloc *entity.Allocator) *Storage {
	return &Storage{
		columns: make(map[component.ID]*column),
		alloc:   alloc,
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
	id := s.alloc.Allocate()
	for _, cv := range components {
		c := s.getOrCreateColumn(cv.ID)
		c.ensureSparseCapacity(id)
		c.sparse[id] = int32(len(c.entities))
		c.entities = append(c.entities, id)
		c.appendValue(cv.Data)
	}
	return id
}

func (s *Storage) Read(id entity.ID, cid component.ID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id entity.ID, cid component.ID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) Query(set []component.ID) storage.Iterator {
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

func (s *Storage) applyDespawn(id entity.ID) {
	for _, c := range s.columns {
		if int(id) < len(c.sparse) && c.sparse[id] != -1 {
			c.swapRemove(id)
		}
	}
	s.alloc.Free(id)
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
