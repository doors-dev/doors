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

type Payload interface {
	Payload() action.Payload
	Release()
}

type PayloadPrinter struct {
	buf     sliceWriter
	gzip    *gzip.Writer
	printer gox.Printer
}

var _ Payload = (*PayloadPrinter)(nil)

func NewPayloadPrinter(disableGzip bool) *PayloadPrinter {
	b := &PayloadPrinter{
		buf: bufferPrinterPool.Get().(sliceWriter),
	}
	if !disableGzip {
		b.gzip = gzip.NewWriter(&b.buf)
		b.printer = defaultPrinter{b.gzip}
	} else {
		b.printer = defaultPrinter{&b.buf}
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
