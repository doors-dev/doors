// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/doors-dev/doors/internal/common"
)

type testInstance struct {
	err error
}

func (t *testInstance) syncError(err error) {
	t.err = err
}

type testCall struct {
	arg       int
	payload   string
	resultErr error
	cancel    bool
}

func (t *testCall) Destroy() {

}

func (t *testCall) Write(w io.Writer) error {
	_, err := w.Write([]byte(t.payload))
	return err
}

func (t *testCall) Data() *common.CallData {
	if t.cancel {
		return nil
	}
	return &common.CallData{
		Name:    "test",
		Arg:     t.arg,
		Payload: t,
	}
}

func (t *testCall) Cancel() {

}

func (t *testCall) Result(r json.RawMessage, err error) {
	t.resultErr = err
}

type testPackage struct {
	isSignal bool
	signal   uint8
	startSeq uint64
	endSeq   uint64
	name     string
	arg      int
	payload  string
}

type testWriter struct {
	buf          bytes.Buffer
	errNextWrite bool
}

func (w *testWriter) readPackage() (*testPackage, error) {
	if w.buf.Len() < 5 {
		return nil, errors.New("not enogh data")
	}
	b, err := w.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x00 {
		return &testPackage{
			isSignal: true,
			signal:   b,
		}, nil
	}
	headerBuf := make([]byte, 0)
	content := false
	for {
		byte, err := w.buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if byte == 0xFF {
			break
		}
		if byte == 0xFC {
			content = true
			break
		}
		headerBuf = append(headerBuf, byte)
	}
	pkg := &testPackage{}
	if content {
		payload, err := w.buf.ReadBytes(0xFF)
		payload = payload[:len(payload)-1]
		if err != nil {
			return nil, err
		}
		if !utf8.Valid(payload) {
			return nil, errors.New("invalid payload encoding")
		}
		pkg.payload = string(payload)
	}
	var header []json.RawMessage
	err = json.Unmarshal(headerBuf, &header)
	if err != nil {
		return nil, err
	}
	if len(header) == 0 || len(header) > 4 {
		return nil, errors.New("wrong sized header")
	}
	err = json.Unmarshal(header[0], &pkg.endSeq)
	if err != nil {
		return nil, err
	}
	pkg.startSeq = pkg.endSeq
	if len(header) == 1 {
		return pkg, nil
	}
	if len(header) == 2 || len(header) == 4 {
		err = json.Unmarshal(header[1], &pkg.startSeq)
		if err != nil {
			return nil, err
		}
	}
	if len(header) == 2 {
		return pkg, nil
	}
	index := 1
	if len(header) == 4 {
		index += 1
	}
	err = json.Unmarshal(header[index], &pkg.name)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(header[index+1], &pkg.arg)
	if err != nil {
		return nil, err
	}
	return pkg, nil

}

func (w *testWriter) Write(b []byte) (int, error) {
	if w.errNextWrite {
		w.errNextWrite = false
		return 0, writerError
	}
	s, _ := w.buf.Write(b)
	return s, nil
}

func (w *testWriter) skip(c int) error {
	for i := range c {
		_, err := w.readPackage()
		if err != nil {
			return errors.Join(err, errors.New("failed"+fmt.Sprint(i)))
		}
	}
	return nil
}
