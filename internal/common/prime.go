// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import (
	"crypto/rand"
	"encoding/binary"
	"math/bits"
	"sync/atomic"
)

const (
	mod        uint64 = (1 << 53) - 1
	multiplier uint64 = 6364136223846793005
	increment  uint64 = 1442695040888963407
)

type Primea struct {
	counter atomic.Uint64
}

func NewPrima() *Primea {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	seed := binary.NativeEndian.Uint64(b[:]) 
	p := &Primea{}
    p.counter.Store(seed)
	return p
}

func (p *Primea) Gen() uint64 {
    c := p.counter.Add(1)
    _, lo := bits.Mul64(c, multiplier)
	lo, _ = bits.Add64(lo, increment, 0)
	return lo & mod
}

