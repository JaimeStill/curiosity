package archetype

import (
	"unsafe"

	"ecs-storage-comparison/storage"
)

type iterator struct {
	matches []*archetype
	index   int
	row     int
}

func (it *iterator) Next() bool {
	it.row++
	for it.index < len(it.matches) && it.row >= len(it.matches[it.index].entities) {
		it.index++
		it.row = 0
	}
	return it.index < len(it.matches)
}

func (it *iterator) Entity() storage.EntityID {
	return it.matches[it.index].entities[it.row]
}

func (it *iterator) Get(cid storage.ComponentID) unsafe.Pointer {
	arch := it.matches[it.index]
	col := arch.columnFor(cid)
	return unsafe.Pointer(&col.data[uintptr(it.row)*col.size])
}
