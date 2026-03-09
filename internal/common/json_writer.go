package common

import "io"

func NewJsonWriter(w io.Writer) io.Writer {
	return jsonWriter{w: w}
}

type jsonWriter struct {
	w io.Writer
}

func (j jsonWriter) Write(b []byte) (n int, err error) {
	adj := 0
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
		adj = 1
	}
	n, err = j.w.Write(b)
	if err != nil {
		return
	}
	n += adj
	return
}
