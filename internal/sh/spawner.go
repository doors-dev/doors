package sh

type Spawner interface {
	Spawn(func())
}

