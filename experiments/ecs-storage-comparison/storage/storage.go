package storage

import (
	"reflect"
	"unsafe"
)

const InvalidComponentID ComponentID = 0

type EntityID uint32
type ComponentID uint16

type ComponentValue struct {
	ID   ComponentID
	Data unsafe.Pointer
}

type Iterator interface {
	Next() bool
	Entity() EntityID
	Get(cid ComponentID) unsafe.Pointer
}

type Storage interface {
	Spawn(components []ComponentValue) EntityID
	Despawn(id EntityID)
	Attach(id EntityID, cid ComponentID, data unsafe.Pointer)
	Detach(id EntityID, cid ComponentID)
	Read(id EntityID, cid ComponentID) (unsafe.Pointer, bool)
	Write(id EntityID, cid ComponentID, data unsafe.Pointer)
	Query(set []ComponentID) Iterator
	ApplyDeferred()
}

var (
	registry = make(map[reflect.Type]ComponentID)
	types    []reflect.Type
)

func ComponentIDFor[T any]() ComponentID {
	t := reflect.TypeFor[T]()
	if id, ok := registry[t]; ok {
		return id
	}
	id := ComponentID(len(types) + 1)
	registry[t] = id
	types = append(types, t)
	return id
}

func ComponentValueFor[T any](value *T) ComponentValue {
	return ComponentValue{ID: ComponentIDFor[T](), Data: unsafe.Pointer(value)}
}

func ComponentType(id ComponentID) reflect.Type {
	if id == InvalidComponentID || int(id) > len(types) {
		return nil
	}
	return types[id-1]
}

func Attach[T any](s Storage, id EntityID, value T) {
	s.Attach(id, ComponentIDFor[T](), unsafe.Pointer(&value))
}

func Detach[T any](s Storage, id EntityID) {
	s.Detach(id, ComponentIDFor[T]())
}

func Read[T any](s Storage, id EntityID) (T, bool) {
	ptr, ok := s.Read(id, ComponentIDFor[T]())
	if !ok {
		var zero T
		return zero, false
	}
	return *(*T)(ptr), true
}

func Write[T any](s Storage, id EntityID, value T) {
	s.Write(id, ComponentIDFor[T](), unsafe.Pointer(&value))
}

func Spawn(s Storage, components ...ComponentValue) EntityID {
	return s.Spawn(components)
}

func Spawn1[A any](s Storage, a A) EntityID {
	return s.Spawn([]ComponentValue{
		ComponentValueFor(&a),
	})
}

func Spawn2[A, B any](s Storage, a A, b B) EntityID {
	return s.Spawn([]ComponentValue{
		ComponentValueFor(&a),
		ComponentValueFor(&b),
	})
}

func Spawn3[A, B, C any](s Storage, a A, b B, c C) EntityID {
	return s.Spawn([]ComponentValue{
		ComponentValueFor(&a),
		ComponentValueFor(&b),
		ComponentValueFor(&c),
	})
}
