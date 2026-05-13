package entity_test

import (
	"testing"

	"github.com/JaimeStill/curiosity/engine/ecs/entity"
)

func TestAllocate_FirstID(t *testing.T) {
	a := entity.New()
	id := a.Allocate()
	if got := id.Index(); got != 1 {
		t.Errorf("Index() = %d, want 1", got)
	}
	if got := id.Generation(); got != 1 {
		t.Errorf("Generation() = %d, want 1", got)
	}
}

func TestAllocate_DistinctIndices(t *testing.T) {
	a := entity.New()
	id1 := a.Allocate()
	id2 := a.Allocate()
	if id1.Index() == id2.Index() {
		t.Errorf("two fresh Allocates produced the same index %d", id1.Index())
	}
}

func TestRecycle_ReusesIndexBumpedGeneration(t *testing.T) {
	a := entity.New()
	id1 := a.Allocate()
	a.Recycle(id1)
	id2 := a.Allocate()
	if id2.Index() != id1.Index() {
		t.Errorf("Index() = %d, want %d", id2.Index(), id1.Index())
	}
	if want := id1.Generation() + 1; id2.Generation() != want {
		t.Errorf("Generation() = %d, want %d", id2.Generation(), want)
	}
}

func TestRecycle_StaleInputNoOp(t *testing.T) {
	a := entity.New()
	id1 := a.Allocate()
	a.Recycle(id1)
	a.Recycle(id1) // stale; silently ignored per the Recycle contract

	id2 := a.Allocate()
	if id2.Index() != id1.Index() {
		t.Errorf("Index() = %d, want %d", id2.Index(), id1.Index())
	}
	if id2.Generation() != 2 {
		t.Errorf("Generation() = %d, want 2", id2.Generation())
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name  string
		setup func() (*entity.Allocator, entity.ID)
		want  bool
	}{
		{
			name: "live id",
			setup: func() (*entity.Allocator, entity.ID) {
				a := entity.New()
				return a, a.Allocate()
			},
			want: true,
		},
		{
			name: "zero id",
			setup: func() (*entity.Allocator, entity.ID) {
				return entity.New(), entity.ID(0)
			},
			want: false,
		},
		{
			name: "recycled id",
			setup: func() (*entity.Allocator, entity.ID) {
				a := entity.New()
				id := a.Allocate()
				a.Recycle(id)
				return a, id
			},
			want: false,
		},
		{
			name: "out-of-range index",
			setup: func() (*entity.Allocator, entity.ID) {
				return entity.New(), entity.ID(uint64(999)<<32 | 1)
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, id := tc.setup()
			if got := a.Validate(id); got != tc.want {
				t.Errorf("Validate(%v) = %v, want %v", id, got, tc.want)
			}
		})
	}
}
