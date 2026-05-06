package sparsesetgroup

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
	group  *group
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

func (it *iterator) Next() bool {
	if len(it.refs) == 0 {
		return false
	}
	if it.group != nil {
		it.row++
		if it.row >= it.group.size {
			return false
		}
		for i := range it.refs {
			it.refs[i].index = it.row
		}
		return true
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
			sparse := it.refs[i].col.sparse
			if int(entity) >= len(sparse) || sparse[entity] < 0 {
				continue NextRow
			}
			it.refs[i].index = int(sparse[entity])
		}
		return true
	}
}
