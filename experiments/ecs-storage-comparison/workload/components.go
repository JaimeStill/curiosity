package workload

type Position struct {
	X, Y, Z float32
}

type Velocity struct {
	X, Y, Z float32
}

type Health struct {
	Current, Max int32
}

type Tag struct {
	V byte
}
