package workload

import (
	"ecs-storage-comparison/component"
	"ecs-storage-comparison/storage"
)

func MultiGroups() [][]component.ID {
	return [][]component.ID{
		{
			component.IDFor[Position](),
			component.IDFor[Velocity](),
			component.IDFor[Health](),
		},
	}
}

func MultiComponentSetup(s storage.Storage, n int) {
	for i := range n {
		switch i % 6 {
		case 0:
			storage.Spawn(
				s,
				component.ValueFor(&Position{X: float32(i)}),
				component.ValueFor(&Velocity{Y: 1}),
				component.ValueFor(&Health{Current: 100, Max: 100}),
				component.ValueFor(&Tag{}),
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
	posID := component.IDFor[Position]()
	velID := component.IDFor[Velocity]()
	healthID := component.IDFor[Health]()
	it := s.Query([]component.ID{posID, velID, healthID})
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
