package sparsesetmap

import (
	"unsafe"

	"ecs-storage-comparison/storage"
)

type queryRef struct {
	col   *column
	index int
}

type iterator struct {
	refs   []queryRef
	driver int
	row    int
}

func (it *iterator) Next() bool {
	if len(it.refs) == 0 {
		return false
	}
	driverCol := it.refs[it.driver].col

NextRow:
	for {
		it.row++
		if it.row >= len(driverCol.entities) {
			return false
		}
		entity := driverCol.entities[it.row]
		for i := range it.refs {
			if i == it.driver {
				it.refs[i].index = it.row
				continue
			}
			index, present := it.refs[i].col.sparse[entity]
			if !present {
				continue NextRow
			}
			it.refs[i].index = index
		}
		return true
	}
}

func (it *iterator) Entity() storage.EntityID {
	return it.refs[it.driver].col.entities[it.row]
}

func (it *iterator) Get(cid storage.ComponentID) unsafe.Pointer {
	for _, ref := range it.refs {
		if ref.col.cid == cid {
			return unsafe.Pointer(
				&ref.col.dense[uintptr(ref.index)*ref.col.size],
			)
		}
	}
	return nil
}
