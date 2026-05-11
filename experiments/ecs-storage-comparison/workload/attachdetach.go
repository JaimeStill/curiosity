package workload

import (
	"ecs-storage-comparison/component"
	"ecs-storage-comparison/entity"
	"ecs-storage-comparison/storage"
)

type attachDetachState struct {
	base      []entity.ID
	decorated []entity.ID
	k         int
}

var attachDetach attachDetachState

func AttachDetachGroups() [][]component.ID {
	return MultiGroups()
}

func AttachDetachSetup(s storage.Storage, n int) {
	half := n / 2
	attachDetach = attachDetachState{
		base:      make([]entity.ID, 0, n),
		decorated: make([]entity.ID, 0, n),
		k:         attachDetachChurnRate(n),
	}
	for i := range half {
		id := storage.Spawn1(s, Position{X: float32(i)})
		attachDetach.base = append(attachDetach.base, id)
	}
	for i := half; i < n; i++ {
		id := storage.Spawn3(s, Position{X: float32(i)}, Velocity{Y: 1}, Health{Current: 100, Max: 100})
		attachDetach.decorated = append(attachDetach.decorated, id)
	}
}

func AttachDetachTick(s storage.Storage) {
	k := attachDetach.k
	for i := range k {
		id := attachDetach.base[i]
		storage.Attach(s, id, Velocity{Y: 1})
		storage.Attach(s, id, Health{Current: 100, Max: 100})
		attachDetach.decorated = append(attachDetach.decorated, id)
	}
	attachDetach.base = attachDetach.base[k:]
	for i := range k {
		id := attachDetach.decorated[i]
		storage.Detach[Velocity](s, id)
		storage.Detach[Health](s, id)
		attachDetach.base = append(attachDetach.base, id)
	}
	attachDetach.decorated = attachDetach.decorated[k:]
	for _, id := range attachDetach.base {
		p, _ := storage.Read[Position](s, id)
		p.X++
		storage.Write(s, id, p)
	}
	for _, id := range attachDetach.decorated {
		p, _ := storage.Read[Position](s, id)
		p.X++
		storage.Write(s, id, p)
	}
	s.ApplyDeferred()
}

func attachDetachChurnRate(n int) int {
	k := n / 100
	if k < 1 {
		return 1
	}
	return k
}
