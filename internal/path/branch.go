// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package path

import (
	"errors"
)

type mutation = func(any) error
type mutations = []mutation

type branch struct {
	atoms  []*atom
	marker *marker
}

func newBranch(path string, marker field) (*branch, error) {
	b := &branch{
		marker: newMarker(marker),
		atoms:  make([]*atom, 0),
	}
	a := newAtom()
	for _, r := range path {
		switch r {
		case ':':
			err := a.capturePart()
			if err != nil {
				return nil, err
			}
		case '|':
			err := a.addTo(b)
			if err != nil {
				return nil, err
			}
			a = newAtom()
		case '+':
			err := a.captureToEnd()
			if err != nil {
				return nil, err
			}
		case '/':
            a.setTail()
			err := a.addTo(b)
			if err != nil {
				return nil, err
			}
			a = newAtom()
		default:
			a.append(r)
		}
	}
    a.setTail()
	a.addTo(b)
	return b, nil
}

func (b *branch) setLastTail() {
	if len(b.atoms) == 0 {
		return
	}
	b.atoms[len(b.atoms)-1].setTail()
}

func (b *branch) setMark(m any) {
    b.marker.set(m)
}

func (b *branch) encode(m any) ([]string, error) {
	parts := make([]string, 0)
	for _, part := range b.atoms {
		p, err := part.encode(m)
		if err != nil {
			return nil, err
		}
		parts = append(parts, p...)
	}
	return parts, nil
}

func (b *branch) checkMark(a any) bool {
	return b.marker.get(a)
}

func (b *branch) decode(p []rune) (mutations, bool) {
	ms := []mutation{b.marker.set}
	if len(b.atoms) == 0 {
		if len(p) == 0 {
			return ms, true
		}
		return nil, false
	}
	m, ok := b.atoms[0].decode(p, b.atoms[1:])
	if ok {
		ms := append(ms, m...)
		return ms, true
	}
	return nil, false
}

func (b *branch) collectParams(s map[string][]*atom) {
	for _, p := range b.atoms {
		p.collectParams(s)
	}
}


func (b *branch) append(part *atom) error {
	if len(b.atoms) > 0 && b.atoms[len(b.atoms)-1].isEnd() {
		return errors.New("Capture syntax error, cannot append after end")
	}
	b.atoms = append(b.atoms, part)
	return nil
}
