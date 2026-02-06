package door

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

type printer struct {
	buf  sliceWriter
	gzip *gzip.Writer
}

func newPrinter(disableGzip bool) *printer {
	b := &printer{
		buf: bufferPrinterPool.Get().(sliceWriter),
	}
	if !disableGzip {
		b.gzip = gzip.NewWriter(&b.buf)
	}
	return b
}

func (b *printer) payload() ([]byte, action.PayloadType) {
	if b == nil {
		return nil, action.PayloadText
	}
	if b.gzip != nil {
		return b.buf, action.PayloadTextGZ
	}
	return b.buf, action.PayloadText
}

func (b *printer) release() {
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

func (b *printer) finalize() {
	if b.gzip == nil {
		return
	}
	b.gzip.Close()
}

func (b *printer) Send(job gox.Job) error {
	if b.gzip != nil {
		return job.Output(b.gzip)
	}
	return job.Output(&b.buf)
}
