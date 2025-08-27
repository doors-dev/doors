// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"unicode"

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
		mu:      sync.Mutex{},
		buffers: make(map[uint64][]byte),
		attrs:   make(map[uint32]*Attrs),
		importMap: &importMap{
			Imports: make(map[string]string),
		},
	}
}

type RenderMap struct {
	mu        sync.Mutex
	buffers   map[uint64][]byte
	attrs     map[uint32]*Attrs
	importMap *importMap
	attrCount uint32
}

func (r *RenderMap) InitImportMap(c *CSPCollector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.importMap.init(c)
}

func (r *RenderMap) AddImport(specifier string, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.importMap.addImport(specifier, path)
}

func (r *RenderMap) WriteImportMap(w io.Writer) error {
	_, err := w.Write([]byte{controlByte, commandImportMap})
	return err
}

func (r *RenderMap) WriteAttrs(w io.Writer, attr *Attrs) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := r.attrCount
	r.attrCount += 1
	r.attrs[index] = attr
	bytes := make([]byte, 6)
	bytes[0] = controlByte
	bytes[1] = commandMagicA
	binary.NativeEndian.PutUint32(bytes[2:], index)
	_, err := w.Write(bytes)
	return err
}

func (r *RenderMap) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buffers = nil
	r.attrs = nil
	r.importMap = nil
}

func (r *RenderMap) RenderBuf(w io.Writer, buf []byte) error {
	return r.renderBuf(w, buf, nil)
}

func (r *RenderMap) Render(w io.Writer, index uint64) error {
	return r.render(w, index, nil)
}

type mode int

const (
	modeLook mode = iota
	modeCommand
	modeInsert
	modeMagicA
	modeMagicAInsert
	modeImportMap
)

const controlByte byte = 0xFF

const (
	commandInsert byte = iota
	commandMagicA
	commandImportMap
)

func (r *RenderMap) renderBuf(w io.Writer, buf []byte, magicA *Attrs) error {
	start := 0
	cursor := 0
	mode := modeLook
	for cursor < len(buf) {
		if mode == modeLook {
			b := buf[cursor]
			if b == controlByte {
				mode = modeCommand
				_, err := w.Write(buf[start:cursor])
				if err != nil {
					return err
				}
				cursor += 1
				continue
			}
			cursor += 1
			if magicA == nil {
				continue
			}
			r := rune(b)
			if unicode.IsSpace(r) {
				continue
			}
			if r == '<' {
				mode = modeMagicAInsert
				continue
			}
			slog.Warn("magic attributes dropped, nowhere to attach")
			magicA = nil
			mode = modeLook
			continue
		}
		if mode == modeCommand {
			command := buf[cursor]
			switch command {
			case commandImportMap:
				mode = modeImportMap
			case commandInsert:
				mode = modeInsert
			case commandMagicA:
				mode = modeMagicA
			default:
				return errors.New("Unsupported command")
			}
			cursor += 1
			continue
		}
		if mode == modeImportMap {
			err := r.importMap.write(w)
			if err != nil {
				return err
			}
			mode = modeLook
			start = cursor
			continue
		}
		if mode == modeInsert {
			if cursor+8 > len(buf) {
				return errors.New("length error")
			}
			id := binary.NativeEndian.Uint64(buf[cursor : cursor+8])
			err := r.render(w, id, magicA)
			magicA = nil
			if err != nil {
				return err
			}
			mode = modeLook
			cursor = cursor + 8
			start = cursor
			continue
		}
		if mode == modeMagicA {
			if cursor+4 > len(buf) {
				return errors.New("length error")
			}
			id := binary.NativeEndian.Uint32(buf[cursor : cursor+4])
			attr, ok := r.attrs[id]
			if !ok {
				return errors.New("magic attr lost")
			}
			delete(r.attrs, id)
			if magicA == nil {
				magicA = attr
			} else {
				magicA.Join(attr)
			}
			cursor = cursor + 4
			mode = modeLook
			start = cursor
			continue
		}
		if mode == modeMagicAInsert {
			r := rune(buf[cursor])
			if !unicode.IsSpace(r) && r != '>' {
				cursor += 1
				continue
			}
			_, err := w.Write(buf[start:cursor])
			if err != nil {
				return err
			}
			err = templ.RenderAttributes(context.Background(), w, magicA)
			if err != nil {
				return err
			}
			magicA = nil
			start = cursor
			mode = modeLook
			continue
		}
	}
	if mode != modeLook {
		return errors.New("buffer missaligned")
	}
	_, err := w.Write(buf[start:])
	return err
}

func (r *RenderMap) render(w io.Writer, index uint64, magicA *Attrs) error {
	r.mu.Lock()
	buf, ok := r.buffers[index]
	if !ok {
		r.mu.Unlock()
		return RenderErrorLog(context.Background(), w, "door "+fmt.Sprint(index)+" not found")
	}
	if buf == nil {
		r.mu.Unlock()
		return RenderErrorLog(context.Background(), w, "door "+fmt.Sprint(index)+" not submitted, bug")
	}
	r.mu.Unlock()
	return r.renderBuf(w, buf, magicA)
}

func (r *RenderMap) Writer(index uint64) (*RenderWriter, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.buffers[index]
	if ok {
		return nil, false
	}
	r.buffers[index] = nil
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
	r.buffers[w.index] = b
}

type RenderWriter struct {
	buf   *bytes.Buffer
	index uint64
	rm    *RenderMap
}

func (rw *RenderWriter) Holdplace(w io.Writer) error {
	bytes := make([]byte, 10)
	bytes[0] = controlByte
	bytes[1] = commandInsert
	binary.NativeEndian.PutUint64(bytes[2:], rw.index)
	_, err := w.Write(bytes)
	return err
}

func (rw *RenderWriter) destroy() {
	rw.buf = nil
	rw.rm = nil
}

func (rw *RenderWriter) Submit() {
	rw.rm.submit(rw)
	rw.destroy()
}

func (rw *RenderWriter) SubmitEmpty() {
	rw.buf.Reset()
	rw.rm.submit(rw)
	rw.destroy()
}
func (rw *RenderWriter) SubmitError(err error, args ...any) {
	rw.buf.Reset()
	RenderErrorLog(context.Background(), rw.buf, err.Error(), args...)
	rw.rm.submit(rw)
	rw.destroy()
}

func (rw *RenderWriter) Len() int {
	return rw.buf.Len()
}

func (w *RenderWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

type importMap struct {
	Imports map[string]string `json:"imports"`
	content []byte
}

func (i *importMap) addImport(specifier string, path string) {
	i.Imports[specifier] = path
}

func (i *importMap) write(w io.Writer) error {
	if i.content == nil {
		return nil
	}
	_, err := w.Write(i.content)
	return err
}

func (i *importMap) init(c *CSPCollector) {
	if len(i.Imports) == 0 {
		return
	}
	w := &bytes.Buffer{}
	json, _ := json.Marshal(i)
	hash := sha256.Sum256(json)
	c.ScriptHash(hash[:])
	w.WriteString("<script type=\"importmap\">")
	w.Write(json)
	w.WriteString("</script>")
	i.content = w.Bytes()
}
