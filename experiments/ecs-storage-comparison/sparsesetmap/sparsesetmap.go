package sparsesetmap

import (
	"unsafe"

	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type column struct {
	cid      component.ID
	size     uintptr
	sparse   map[entity.ID]int
	dense    []byte
	entities []entity.ID
}

func (c *column) appendValue(src unsafe.Pointer) {
	c.dense = append(c.dense, unsafe.Slice((*byte)(src), c.size)...)
}

type Storage struct {
	columns map[component.ID]*column
	alloc   *entity.Allocator
}

var _ storage.Storage = (*Storage)(nil)

func New(alloc *entity.Allocator) *Storage {
	return &Storage{
		columns: make(map[component.ID]*column),
		alloc:   alloc,
	}
}

func (s *Storage) ApplyDeferred() {
	panic("not implemented")
}

func (s *Storage) Attach(id entity.ID, cid component.ID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) Detach(id entity.ID, cid component.ID) {
	panic("not implemented")
}

func (s *Storage) Despawn(id entity.ID) {
	panic("not implemented")
}

func (s *Storage) Spawn(components []component.Value) entity.ID {
	id := s.alloc.Allocate()
	for _, cv := range components {
		c := s.getOrCreateColumn(cv.ID)
		c.sparse[id] = len(c.entities)
		c.entities = append(c.entities, id)
		c.appendValue(cv.Data)
	}
	return id
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

func (s *Storage) Read(id entity.ID, cid component.ID) (unsafe.Pointer, bool) {
	panic("not implemented")
}

func (s *Storage) Write(id entity.ID, cid component.ID, data unsafe.Pointer) {
	panic("not implemented")
}

func (s *Storage) getOrCreateColumn(cid component.ID) *column {
	if c, ok := s.columns[cid]; ok {
		return c
	}
	c := &column{
		cid:    cid,
		size:   component.TypeOf(cid).Size(),
		sparse: make(map[entity.ID]int),
	}
	s.columns[cid] = c
	return c
}
