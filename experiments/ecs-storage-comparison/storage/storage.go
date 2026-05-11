package storage

import (
	"unsafe"

	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
)

type Iterator interface {
	Next() bool
	Entity() entity.ID
	Get(cid component.ID) unsafe.Pointer
}

type Storage interface {
	Attach(id entity.ID, cid component.ID, data unsafe.Pointer)
	Detach(id entity.ID, cid component.ID)
	Spawn(components []component.Value) entity.ID
	Despawn(id entity.ID)
	Query(set []component.ID) Iterator
	Read(id entity.ID, cid component.ID) (unsafe.Pointer, bool)
	Write(id entity.ID, cid component.ID, data unsafe.Pointer)
	ApplyDeferred()
}

func Attach[T any](s Storage, id entity.ID, value T) {
	s.Attach(id, component.IDFor[T](), unsafe.Pointer(&value))
}

func Detach[T any](s Storage, id entity.ID) {
	s.Detach(id, component.IDFor[T]())
}

func Read[T any](s Storage, id entity.ID) (T, bool) {
	ptr, ok := s.Read(id, component.IDFor[T]())
	if !ok {
		var zero T
		return zero, false
	}
	return *(*T)(ptr), true
}

func Write[T any](s Storage, id entity.ID, value T) {
	s.Write(id, component.IDFor[T](), unsafe.Pointer(&value))
}

func Spawn(s Storage, components ...component.Value) entity.ID {
	return s.Spawn(components)
}

func Spawn1[A any](s Storage, a A) entity.ID {
	return s.Spawn([]component.Value{
		component.ValueFor(&a),
	})
}

func Spawn2[A, B any](s Storage, a A, b B) entity.ID {
	return s.Spawn([]component.Value{
		component.ValueFor(&a),
		component.ValueFor(&b),
	})
}

func Spawn3[A, B, C any](s Storage, a A, b B, c C) entity.ID {
	return s.Spawn([]component.Value{
		component.ValueFor(&a),
		component.ValueFor(&b),
		component.ValueFor(&c),
	})
}
