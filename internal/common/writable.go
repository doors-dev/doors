package common

import (
	"encoding/json"
	"io"
	"log"
)

type Writable interface {
    Destroy() 
	WriteJson(io.Writer) error
}

var bo = []byte("[")
var bc = []byte("]")
var comma = []byte(",")



type WritableAny struct {
	V any
}

func (wa WritableAny) Destroy() {

}

func (wa WritableAny) WriteJson(w io.Writer) error {
	buf, err := json.Marshal(wa.V)
    if err != nil {
        log.Fatalf("Can't marshal")
    }
    _, err = w.Write(buf)
	return err
}

type Writables []Writable

func (a Writables) Destroy() {
    for _, writable := range a {
        writable.Destroy()
    }
}
func (a Writables) WriteJson(w io.Writer) error {
	_, err := w.Write(bo)
	if err != nil {
		return err
	}
	for i, writable := range a {
		err = writable.WriteJson(w)
		if err != nil {
			return err
		}
		if i+1 == len(a) {
			break
		}
		_, err = w.Write(comma)
		if err != nil {
			return err
		}
	}
	_, err = w.Write(bc)
	return err
}

type WritableRenderMap struct {
	Rm    *RenderMap
	Index uint64
}


func (wrm *WritableRenderMap) Destroy() {
    wrm.Rm.Destroy()
}
func (wrm *WritableRenderMap) WriteJson(w io.Writer) error {
	return wrm.Rm.RenderJson(w, wrm.Index)
}

type WritableRaw []byte

func (wr WritableRaw) Destroy() {
}
func (wr WritableRaw) WriteJson(w io.Writer) error {
    _, err := w.Write(([]byte)(wr))
    return err
}

