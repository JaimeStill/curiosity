package component

import "unsafe"

// ID identifies a component type within an ECS world. IDs are uint16 —
// sized to pair with archetype storage's flat-array column index per
// D-030 §6 — and are assigned by IDFor on first reference to a Go type.
// Valid IDs start at 1 and grow contiguously. The zero value (InvalidID)
// is intrinsically invalid: it is the sentinel for "no component" and is
// never produced by the registry.
type ID uint16

// InvalidID is the zero ID, reserved as the sentinel for "no component."
// It is never assigned to a Go type by IDFor.
const InvalidID ID = 0

// Value is the type-erased component carrier used by ECS internals: an
// ID identifying the component type plus an unsafe.Pointer to the value's
// storage. Value is contained to the package — it is constructed by
// internal helpers and consumed by archetype storage; the typed
// call-site surface (Attach, Set, Get) never exposes Value to user code.
//
// The unsafe.Pointer carrier is an inner-tier divergence per conventions
// §2: dispatching the hot path through reflect would dominate per-call
// cost. The unsafe.Pointer is contained to this single carrier type.
type Value struct {
	ID   ID
	Data unsafe.Pointer
}
