// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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

type Prime struct {
	counter atomic.Uint64
}

func NewPrime() *Prime {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	seed := binary.NativeEndian.Uint64(b[:])
	p := &Prime{}
	p.counter.Store(seed)
	return p
}

func (p *Prime) Gen() uint64 {
	c := p.counter.Add(1)
	_, lo := bits.Mul64(c, multiplier)
	lo, _ = bits.Add64(lo, increment, 0)
	num := lo & mod
	if num == 0 {
		return p.Gen()
	}
	return num
}
