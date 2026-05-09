package entity

type ID uint32

type Allocator struct {
	free []ID
	next ID
}

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (a *Allocator) Allocate() ID {
	if n := len(a.free); n > 0 {
		id := a.free[n-1]
		a.free = a.free[:n-1]
		return id
	}
	a.next++
	return a.next
}

func (a *Allocator) Free(id ID) {
	a.free = append(a.free, id)
}
