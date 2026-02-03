package common

import (
	"compress/gzip"
	"sync"

	"github.com/doors-dev/gox"
)

var bufferPrinterPool = sync.Pool{
	New: func() any {
		return sliceWriter(make([]byte, 0, 1024))
	},
}

type sliceWriter []byte

func (s *sliceWriter) Write(p []byte) (n int, err error) {
	*s = append(*s, p...)
	return len(p), nil
}

type BufferPrinter struct {
	buf  sliceWriter
	gzip *gzip.Writer
}

func NewBufferPrinter(disableGzip bool) *BufferPrinter {
	b := &BufferPrinter{
		buf: bufferPrinterPool.Get().(sliceWriter),
	}
	if !disableGzip {
		b.gzip = gzip.NewWriter(&b.buf)
	}
	return b
}

func (b *BufferPrinter) Bytes() []byte {
	if b == nil {
		return nil
	}
	return b.buf
}

func (b *BufferPrinter) Release() {
	if b == nil {
		return
	}
	if b.buf == nil {
		return
	}
	b.buf = b.buf[:0]
	bufferPrinterPool.Put(b.buf)
	b.buf = nil
}

func (b *BufferPrinter) Finalize() {
	if b.gzip == nil {
		return
	}
	b.gzip.Close()
}

func (b *BufferPrinter) Send(job gox.Job) error {
	if b.gzip != nil {
		return job.Output(b.gzip)
	}
	return job.Output(&b.buf)
}
