package entity

// Allocator hands out entity IDs and tracks their generations. Each
// World owns one allocator; the allocator is not a package-level
// singleton. Indices are recycled via a FIFO queue so the gap
// between reuses of the same slot stays as wide as possible,
// maximizing the generation-spread that protects against ABA
// aliasing.
type Allocator struct {
	generations []uint32
	recycle     []uint32
	head        int
}

// New returns a fresh Allocator. Index 0 is reserved at construction
// time so the zero ID is always invalid.
func New() *Allocator {
	return &Allocator{
		generations: []uint32{0},
	}
}

// Allocate returns a fresh or recycled ID. When the recycle queue
// is non-empty, the next-queued index is reused with the generation
// previously bumped at Recycle time. Otherwise a new index is
// allocated at the high end of the generation table with initial
// generation 1.
func (a *Allocator) Allocate() ID {
	if a.head < len(a.recycle) {
		idx := a.recycle[a.head]
		a.head++
		if a.head == len(a.recycle) {
			a.recycle = a.recycle[:0]
			a.head = 0
		}
		return newID(idx, a.generations[idx])
	}
	idx := uint32(len(a.generations))
	a.generations = append(a.generations, 1)
	return newID(idx, 1)
}

// Recycle returns id's index to the recycle queue and bumps its
// generation. Stale and out-of-range IDs are silently ignored —
// duplicate recycles during a single ApplyDeferred pass are
// legitimate and require no error. Generation 0 is skipped on wrap
// so the allocator's "never produces generation 0" invariant holds.
func (a *Allocator) Recycle(id ID) {
	idx := id.Index()
	if int(idx) >= len(a.generations) {
		return
	}
	if a.generations[idx] != id.Generation() {
		return
	}
	next := a.generations[idx] + 1
	if next == 0 {
		next = 1
	}
	a.generations[idx] = next
	a.recycle = append(a.recycle, idx)
}

// Validate reports whether id is currently live: its index is in
// range, its generation matches the table, and it is not the zero
// ID. World-level operations translate a false result into
// ErrStaleEntity at their boundary.
func (a *Allocator) Validate(id ID) bool {
	idx := id.Index()
	if idx == 0 || int(idx) >= len(a.generations) {
		return false
	}
	return a.generations[idx] == id.Generation()
}
