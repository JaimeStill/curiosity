package workload

import (
	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type mixedState struct {
	base      []entity.ID
	decorated []entity.ID
	k         int
}

var mixed mixedState

func MixedGroups() [][]component.ID {
	return MultiGroups()
}

func MixedSetup(s storage.Storage, n int) {
	half := n / 2
	mixed = mixedState{
		base:      make([]entity.ID, 0, n),
		decorated: make([]entity.ID, 0, n),
		k:         mixedChurnRate(n),
	}
	for i := range half {
		id := storage.Spawn2(s, Position{X: float32(i)}, Velocity{Y: 1})
		mixed.base = append(mixed.base, id)
	}
	for i := half; i < n; i++ {
		id := storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
		mixed.decorated = append(mixed.decorated, id)
	}
}

func MixedTick(s storage.Storage) {
	posID := component.IDFor[Position]()
	velID := component.IDFor[Velocity]()
	healthID := component.IDFor[Health]()

	it := s.Query([]component.ID{posID, velID})
	for it.Next() {
		pos := (*Position)(it.Get(posID))
		vel := (*Velocity)(it.Get(velID))
		pos.X += vel.X
		pos.Y += vel.Y
		pos.Z += vel.Z
	}

	it = s.Query([]component.ID{posID, velID, healthID})
	for it.Next() {
		health := (*Health)(it.Get(healthID))
		health.Current -= 1
	}

	k := mixed.k

	for i := range k {
		s.Despawn(mixed.base[i])
	}
	mixed.base = mixed.base[k:]

	for i := range k {
		id := storage.Spawn2(s, Position{X: float32(i)}, Velocity{Y: 1})
		mixed.base = append(mixed.base, id)
	}

	for i := range k {
		id := mixed.base[i]
		storage.Attach(s, id, Health{Current: 100, Max: 100})
		mixed.decorated = append(mixed.decorated, id)
	}
	mixed.decorated = mixed.decorated[k:]

	s.ApplyDeferred()
}

func mixedChurnRate(n int) int {
	k := n / 100
	if k < 1 {
		return 1
	}
	return k
}
