package door2

import "github.com/doors-dev/doors/internal/beam2"

type Core = *core

func newCore(cinema beam2.Cinema) Core {
	return &core{
		cinema: cinema,
	}
}

type core struct {
	cinema beam2.Cinema
}

func (c Core) Cinema() beam2.Cinema {
	return c.cinema
}

var _ beam2.Core = &core{}
