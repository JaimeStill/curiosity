package workload

import "ecs-storage-comparison/storage"

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
	posID := storage.ComponentIDFor[Position]()
	velID := storage.ComponentIDFor[Velocity]()
	it := s.Query([]storage.ComponentID{posID, velID})
	for it.Next() {
		pos := (*Position)(it.Get(posID))
		vel := (*Velocity)(it.Get(velID))
		pos.X += vel.X
		pos.Y += vel.Y
		pos.Z += vel.Z
	}
}
