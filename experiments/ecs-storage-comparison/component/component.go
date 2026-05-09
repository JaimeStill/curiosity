package component

import (
	"reflect"
	"unsafe"
)

type ID uint16

const InvalidID ID = 0

type Value struct {
	ID   ID
	Data unsafe.Pointer
}

var (
	registry = make(map[reflect.Type]ID)
	types    []reflect.Type
)

func IDFor[T any]() ID {
	t := reflect.TypeFor[T]()
	if id, ok := registry[t]; ok {
		return id
	}
	id := ID(len(types) + 1)
	registry[t] = id
	types = append(types, t)
	return id
}

func ValueFor[T any](value *T) Value {
	return Value{ID: IDFor[T](), Data: unsafe.Pointer(value)}
}

func TypeOf(id ID) reflect.Type {
	if id == InvalidID || int(id) > len(types) {
		return nil
	}
	return types[id-1]
}
