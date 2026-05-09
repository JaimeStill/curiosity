package workload

import (
	"ecs-storage-comparison/component"
	"ecs-storage-comparison/storage"
)

func IterationGroups() [][]component.ID {
	return [][]component.ID{
		{
			component.IDFor[Position](),
			component.IDFor[Velocity](),
		},
	}
}

func IterationSetup(s storage.Storage, n int) {
	for i := range n {
		storage.Spawn2(
			s,
			Position{X: float32(i)},
			Velocity{Y: 1},
		)
	}
}

func IterationTick(s storage.Storage) {
	posID := component.IDFor[Position]()
	velID := component.IDFor[Velocity]()
	it := s.Query([]component.ID{posID, velID})
	for it.Next() {
		pos := (*Position)(it.Get(posID))
		vel := (*Velocity)(it.Get(velID))
		pos.X += vel.X
		pos.Y += vel.Y
		pos.Z += vel.Z
	}
}
