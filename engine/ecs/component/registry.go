package component

import "reflect"

// registry and types together form the process-wide component-ID
// registry. The registry map is IDFor's cache; the types slice is the
// source of truth, indexed by id-1 (since InvalidID = 0 consumes the
// zero position).
//
// Process-wide is deliberate: component identity must be stable across
// every World in the process so signatures, query views, and archetype
// layouts are comparable. This is the counterpoint to entity.Allocator,
// which is per-World.
//
// Both are unsynchronized — an inner-tier divergence per conventions §2.
// D-024 settled single-threaded execution, and ID assignment is one-time
// per T and amortizes to zero after the first call. Revisit when
// threading evolves (concepts/engine/scheduler.md).
var (
	registry = make(map[reflect.Type]ID)
	types    []reflect.Type
)

// IDFor returns the component ID for T, assigning a fresh one on first
// reference. The mapping from Go type to ID is process-wide and stable:
// every World in the process sees a given T as the same ID. Calls are
// cached after the first; subsequent calls for the same T are a single
// map lookup.
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

// TypeOf returns the reflect.Type registered to id, or nil if id is
// InvalidID or has not been assigned. It is the reverse direction of
// IDFor.
func TypeOf(id ID) reflect.Type {
	if id == InvalidID || int(id) > len(types) {
		return nil
	}
	return types[id-1]
}
