package printer

import (
	"compress/gzip"
	"sync"

	"github.com/doors-dev/doors/internal/front/action"
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

type PayloadPrinter struct {
	buf     sliceWriter
	gzip    *gzip.Writer
	printer gox.Printer
}

func NewPayloadPrinter(disableGzip bool) *PayloadPrinter {
	b := &PayloadPrinter{
		buf: bufferPrinterPool.Get().(sliceWriter),
	}
	if !disableGzip {
		b.gzip = gzip.NewWriter(&b.buf)
		b.printer = newResourcePrinter(defaultPrinter{b.gzip})
	} else {
		b.printer = newResourcePrinter(defaultPrinter{&b.buf})
	}
	return b
}

func (b *PayloadPrinter) Payload() action.Payload {
	if b.gzip != nil {
		return action.NewTextGZ(b.buf)
	}
	return action.NewTextBytes(b.buf)
}

func (b *PayloadPrinter) Release() {
	if b.buf == nil {
		return
	}
	b.buf = b.buf[:0]
	bufferPrinterPool.Put(b.buf)
	b.buf = nil
}

func (b *PayloadPrinter) Finalize() {
	if b.gzip == nil {
		return
	}
	b.gzip.Close()
}

func (b *PayloadPrinter) Send(job gox.Job) error {
	return b.printer.Send(job)
}
