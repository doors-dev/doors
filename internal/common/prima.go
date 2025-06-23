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

type Prima struct {
	counter atomic.Uint64
}

func NewPrima() *Prima {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	seed := binary.NativeEndian.Uint64(b[:]) 
	p := &Prima{}
    p.counter.Store(seed)
	return p
}

func (p *Prima) Gen() uint64 {
    c := p.counter.Add(1)
    _, lo := bits.Mul64(c, multiplier)
	lo, _ = bits.Add64(lo, increment, 0)
	return lo & mod
}

