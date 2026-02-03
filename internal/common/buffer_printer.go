package common

import (
	"sync"

	"github.com/doors-dev/gox"
)

var bufferPrinterPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 1024)
	},
}

type BufferPrinter []byte

func NewBufferPrinter() *BufferPrinter {
	b := bufferPrinterPool.Get().([]byte)[:0]
	bp := BufferPrinter(b)
	return &bp
}

func (b *BufferPrinter) Bytes() []byte {
	if b == nil {
		return nil
	}
	return *b
}

func (b *BufferPrinter) Release() {
	if b == nil {
		return
	}
	bytes := ([]byte)(*b)[:0]
	bufferPrinterPool.Put(bytes)
	*b = nil
}

func (b *BufferPrinter) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

func (b *BufferPrinter) Send(job gox.Job) error {
	return job.Output(b)
}
