package common

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/a-h/templ"
)

func RenderErrorLog(ctx context.Context, w io.Writer, message string, args ...any) error {
	slog.Error("Render: "+message, args...)
	return RenderError(message).Render(ctx, w)
}

func RenderError(msg string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div style='background:red;color:black;'><strong>RENDER ERROR:</strong> "))
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("</div>"))
		return err
	})
}

func NewRenderMap() *RenderMap {
	return &RenderMap{
		mu: sync.Mutex{},
		m:  make(map[uint64][]byte),
	}
}

type RenderMap struct {
	mu sync.Mutex
	m  map[uint64][]byte
}

func (r *RenderMap) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m = nil
}

func (r *RenderMap) RenderBuf(w io.Writer, buf []byte) error {
	return r.renderBuf(w, buf, false)
}

func (r *RenderMap) Render(w io.Writer, index uint64) error {
	return r.render(w, index, false)
}

var qoute = []byte(`"`)

func (r *RenderMap) RenderJsonBuf(w io.Writer, buf []byte) error {
	_, err := w.Write(qoute)
	if err != nil {
		return err
	}
	err = r.renderBuf(w, buf, true)
	if err != nil {
		return err
	}
	_, err = w.Write(qoute)
	return err
}

func (r *RenderMap) RenderJson(w io.Writer, index uint64) error {
	_, err := w.Write(qoute)
	if err != nil {
		return err
	}
	err = r.render(w, index, true)
	if err != nil {
		return err
	}
	_, err = w.Write(qoute)
	return err
}

var escape = map[byte][]byte{
	'"':  []byte("\\\""),
	'\\': []byte("\\\\"),
	'/':  []byte("\\/"),
	'\b': []byte("\\b"),
	'\f': []byte("\\f"),
	'\n': []byte("\\n"),
	'\r': []byte("\\r"),
	'\t': []byte("\\t"),
}

func (r *RenderMap) renderBuf(w io.Writer, buf []byte, json bool) error {
	nextIndexCursor := -1
	nextIndexBuffer := make([]byte, 8)
	start := 0
	for i, byte := range buf {
		if nextIndexCursor != -1 {
			nextIndexBuffer[nextIndexCursor] = byte
			nextIndexCursor += 1
			if nextIndexCursor == 8 {
				nextIndexCursor = -1
				start = i + 1
				err := r.render(w, binary.NativeEndian.Uint64(nextIndexBuffer), json)
				if err != nil {
					return err
				}
			}
			continue
		}
		if byte == 0xFF {
			nextIndexCursor = 0
			_, err := w.Write(buf[start:i])
			if err != nil {
				return err
			}
			continue
		}
		if !json {
			continue
		}
		escape, ok := escape[byte]
		if ok {
			_, err := w.Write(buf[start:i])
			if err != nil {
				return err
			}
			_, err = w.Write(escape)
			if err != nil {
				return err
			}
			start = i + 1
			continue
		}
		if byte < 32 {
			_, err := w.Write(buf[start:i])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "\\u%04x", byte)
			if err != nil {
				return err
			}
			start = i + 1
		}

	}
	_, err := w.Write(buf[start:])
	return err
}

func (r *RenderMap) render(w io.Writer, index uint64, json bool) error {
	r.mu.Lock()
	buf, ok := r.m[index]
	if !ok {
		r.mu.Unlock()
		return RenderErrorLog(context.Background(), w, "node "+fmt.Sprint(index)+" not found")
	}
	if buf == nil {
		r.mu.Unlock()
		return RenderErrorLog(context.Background(), w, "node "+fmt.Sprint(index)+" not submitted, bug")
	}
	r.mu.Unlock()
	return r.renderBuf(w, buf, json)
}

func (r *RenderMap) Writer(index uint64) (*RenderWriter, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.m[index]
	if ok {
		return nil, false
	}
	r.m[index] = nil
	return &RenderWriter{
		index: index,
		buf:   &bytes.Buffer{},
		rm:    r,
	}, true
}

func (r *RenderMap) submit(w *RenderWriter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := w.buf.Bytes()
	if b == nil {
		b = make([]byte, 0)
	}
	r.m[w.index] = b
}

type RenderWriter struct {
	buf   *bytes.Buffer
	index uint64
	rm    *RenderMap
}

func (rw *RenderWriter) Holdplace(w io.Writer) error {
	bytes := make([]byte, 9)
	bytes[0] = 0xFF
	binary.NativeEndian.PutUint64(bytes[1:], rw.index)
	_, err := w.Write(bytes)
	return err
}

func (rw *RenderWriter) Submit() {
	rw.rm.submit(rw)
}

func (rw *RenderWriter) SubmitEmpty() {
	rw.buf.Reset()
	rw.rm.submit(rw)
}
func (rw *RenderWriter) SubmitError(err error, args ...any) {
	rw.buf.Reset()
	RenderErrorLog(context.Background(), rw.buf, err.Error(), args...)
	rw.rm.submit(rw)
}

func (rw *RenderWriter) Len() int {
	return rw.buf.Len()
}

func (w *RenderWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
