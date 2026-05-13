package entity

// ID identifies an entity in an ECS world. It is a uint64 packed
// with a 32-bit index in the high half and a 32-bit generation in
// the low half. The zero value (ID(0)) is intrinsically invalid:
// the Allocator reserves index 0 and never produces generation 0.
//
// IDs are obtained exclusively through Allocator.Allocate; there is
// no public constructor that takes (index, generation). This
// structurally enforces the invariant that every live ID traces
// back to an allocator that knows its generation.
type ID uint64

// Index returns the entity's index — the slot the allocator
// assigned it within the generation table.
func (id ID) Index() uint32 {
	return uint32(id >> 32)
}

// Generation returns the entity's generation — the recycle counter
// for its index. Combined with Index, it uniquely identifies the
// entity across the program's lifetime (until the generation
// counter wraps, which is unreachable at voxel-game scale).
func (id ID) Generation() uint32 {
	return uint32(id)
}

func newID(index, generation uint32) ID {
	return ID(uint64(index)<<32 | uint64(generation))
}
