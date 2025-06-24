package common

import (
	"encoding/json"
	"io"
	"log"
)

type Writable interface {
	Destroy()
	Write(io.Writer) error
}

type JsonWritable interface {
	WriteJson(io.Writer) error
}

var bo = []byte("[")
var bc = []byte("]")
var comma = []byte(",")

type JsonWritableAny struct {
	V any
}

func (wa JsonWritableAny) WriteJson(w io.Writer) error {
	buf, err := json.Marshal(wa.V)
	if err != nil {
		log.Fatalf("Can't marshal")
	}
	_, err = w.Write(buf)
	return err
}

type JsonWritables []JsonWritable

func (a JsonWritables) WriteJson(w io.Writer) error {
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
func (wrm *WritableRenderMap) Write(w io.Writer) error {
	return wrm.Rm.Render(w, wrm.Index)
}

type JsonWritabeRaw []byte

func (j JsonWritabeRaw) WriteJson(w io.Writer) error {
	_, err := w.Write(j)
	return err
}



