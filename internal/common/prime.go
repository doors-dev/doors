// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type Prime = *prime

type prime struct {
	counter atomic.Uint64
}

func NewPrime() Prime {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	seed := binary.NativeEndian.Uint64(b[:])
	p := &prime{}
	p.counter.Store(seed)
	return p
}

func (p Prime) Gen() uint64 {
	c := p.counter.Add(1)
	_, lo := bits.Mul64(c, multiplier)
	lo, _ = bits.Add64(lo, increment, 0)
	num := lo & mod
	if num == 0 {
		return p.Gen()
	}
	return num
}
