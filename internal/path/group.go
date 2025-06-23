package path

import (
	"errors"
)

type group struct {
	branches []*branch
}

func newGroup() *group {
	return &group{
		branches: make([]*branch, 0),
	}
}

func (g *group) encode(m any) ([]string, error) {
	if len(g.branches) == 1 {
		g.branches[0].setMark(m)
	}
	for _, b := range g.branches {
		if b.checkMark(m) {
			return b.encode(m)
		}
	}
	return nil, errors.New("Could not match any branch against existing markers")
}

func (g *group) decode(p []rune) (mutations, bool) {
	for _, b := range g.branches {
		mutations, ok := b.decode(p)
		if ok {
			return mutations, true
		}
	}
	return nil, false
}

func (g *group) collectParams(s map[string][]*atom) {
	for _, b := range g.branches {
		b.collectParams(s)
	}
}

func (g *group) append(branch *branch) error {
	g.branches = append(g.branches, branch)
	return nil
}
