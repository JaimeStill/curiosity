package workload

import "ecs-storage-comparison/storage"

func MultiGroups() [][]storage.ComponentID {
	return [][]storage.ComponentID{
		{
			storage.ComponentIDFor[Position](),
			storage.ComponentIDFor[Velocity](),
			storage.ComponentIDFor[Health](),
		},
	}
}

func MultiComponentSetup(s storage.Storage, n int) {
	for i := range n {
		switch i % 6 {
		case 0:
			storage.Spawn(
				s,
				storage.ComponentValueFor(&Position{X: float32(i)}),
				storage.ComponentValueFor(&Velocity{Y: 1}),
				storage.ComponentValueFor(&Health{Current: 100, Max: 100}),
				storage.ComponentValueFor(&Tag{}),
			)
		case 1:
			storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
		case 2:
			storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Tag{})
		case 3:
			storage.Spawn2(s, Position{X: float32(i)}, Velocity{Y: 1})
		case 4:
			storage.Spawn3(s, Position{X: float32(i)}, Health{Current: 100, Max: 100}, Tag{})
		case 5:
			storage.Spawn2(s, Velocity{Y: 1}, Tag{})
		}
	}
}

func MultiFullTick(s storage.Storage) {
	posID := storage.ComponentIDFor[Position]()
	velID := storage.ComponentIDFor[Velocity]()
	healthID := storage.ComponentIDFor[Health]()
	it := s.Query([]storage.ComponentID{posID, velID, healthID})
	for it.Next() {
		pos := (*Position)(it.Get(posID))
		vel := (*Velocity)(it.Get(velID))
		health := (*Health)(it.Get(healthID))
		pos.X += vel.X
		pos.Y += vel.Y
		pos.Z += vel.Z
		health.Current -= 1
	}
}

func MultiPartialTick(s storage.Storage) {
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
