package common

import (
	"io"
)

type WritableNone struct{}

func (w WritableNone) Destroy() {

}

func (w WritableNone) Write(io.Writer) error {
	return nil
}

type Writable interface {
	Destroy()
	Write(io.Writer) error
}

type WritableRenderMap struct {
	Rm    *RenderMap
	Index uint64
}

func (wrm *WritableRenderMap) Destroy() {
	wrm.Rm.Destroy()
}
func (wrm *WritableRenderMap) Write(w io.Writer) error {
	return wrm.Rm.Render(w, wrm.Index)
}
