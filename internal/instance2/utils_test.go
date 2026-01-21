// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
)

type testInstance struct {
	err error
}

func (t *testInstance) touch() {
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

func (y *testCall) Params() action.CallParams {
	return action.CallParams{}
}

func (t *testCall) Write(w io.Writer) error {
	_, err := w.Write([]byte(t.payload))
	return err
}

func (t *testCall) Payload() common.Writable {
	return t
}

func (t *testCall) Action() (action.Action, bool) {
	if t.cancel {
		return nil, false
	}
	return &action.Test{
		Arg: t.arg,
	}, true
}

func (t *testCall) Clean() {

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

func (*testWriter) afterFlush(f func()) {
	f()
}
func (w *testWriter) readPackage() (*testPackage, error) {
	if w.buf.Len() < 5 {
		return nil, errors.New("not enogh data")
	}
	b, err := w.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
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
	if len(header) == 0 || len(header) > 2 {
		return nil, errors.New("wrong sized header")
	}
	rg := make([]uint64, 0)
	err = json.Unmarshal(header[0], &rg)
	if err != nil {
		return nil, err
	}
	pkg.endSeq = rg[0]
	if len(rg) == 2 {
		pkg.startSeq = rg[1]
	} else {
		pkg.startSeq = pkg.endSeq
	}
	if len(header) == 1 {
		return pkg, nil
	}
	var inv [2]json.RawMessage
	err = json.Unmarshal(header[1], &inv)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inv[0], &pkg.name)
	if err != nil {
		return nil, err
	}
	var arg [1]int
	err = json.Unmarshal(inv[1], &arg)
	if err != nil {
		return nil, err
	}
	pkg.arg = arg[0]
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
