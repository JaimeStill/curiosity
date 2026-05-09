package workload

import (
	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type structuralState struct {
	alive   []entity.ID
	head    int
	k       int
	target  int
	growing bool
}

var structural structuralState

func StructuralGroups() [][]component.ID {
	return MultiGroups()
}

func StructuralCycleSetup(s storage.Storage, n int) {
	structural = structuralState{
		alive:   make([]entity.ID, 0, n),
		k:       structuralChurnRate(n),
		target:  n,
		growing: true,
	}
}

func StructuralCycleTick(s storage.Storage) {
	k := structural.k
	if structural.growing {
		for i := range k {
			id := storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
			structural.alive = append(structural.alive, id)
		}
		if len(structural.alive)-structural.head >= structural.target {
			structural.growing = false
		}
	} else {
		for i := range k {
			s.Despawn(structural.alive[structural.head+i])
		}
		structural.head += k
		if structural.head >= len(structural.alive) {
			structural.growing = true
			structural.alive = structural.alive[:0]
			structural.head = 0
		}
	}
	s.ApplyDeferred()
}

func StructuralGrowthSetup(s storage.Storage, n int) {
	structural = structuralState{
		k: structuralChurnRate(n),
	}
}

func StructuralGrowthTick(s storage.Storage) {
	k := structural.k
	for i := range k {
		storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
	}
	s.ApplyDeferred()
}

func StructuralSteadySetup(s storage.Storage, n int) {
	structural = structuralState{
		alive: make([]entity.ID, 0, n),
		k:     structuralChurnRate(n),
	}
	for i := range n {
		id := storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
		structural.alive = append(structural.alive, id)
	}
}

func StructuralSteadyTick(s storage.Storage) {
	k := structural.k
	for i := range k {
		s.Despawn(structural.alive[structural.head+i])
	}
	structural.head += k
	for i := range k {
		id := storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
		structural.alive = append(structural.alive, id)
	}
	s.ApplyDeferred()
}

func structuralChurnRate(n int) int {
	k := n / 100
	if k < 1 {
		return 1
	}
	return k
}
