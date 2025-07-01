package instance

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
)

type testCall struct {
	arg       int
	payload   string
	resultErr error
	cancel    bool
}

func (t *testCall) Destroy() {}


func (t *testCall) Write(w io.Writer) error {
	_, err := w.Write([]byte(t.payload))
	return err
}

func (t *testCall) Payload() (common.Writable, bool) {
	if t.payload == "" {
		return nil, false
	}
	return t, true
}

func (t *testCall) Call() (node.Call, bool) {
	return t, !t.cancel
}

func (t *testCall) OnWriteErr() bool {
	return !t.cancel
}

func (t *testCall) OnResult(err error) {
	t.resultErr = err
}

func (t *testCall) Name() string {
	return "test"
}

func (t *testCall) Arg() common.JsonWritable {
	return common.JsonWritableAny{t.arg}
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
	var headerLength uint32
	err := binary.Read(&w.buf, binary.BigEndian, &headerLength)
	if err != nil {
		return nil, err
	}
	if headerLength == 0 {
		b, err := w.buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return &testPackage{
			isSignal: true,
			signal:   b,
		}, nil
	}
	buf := make([]byte, headerLength)
	_, err = w.buf.Read(buf)
	if err != nil {
		return nil, err
	}
	payload, err := w.buf.ReadBytes(0xFF)
	payload = payload[:len(payload)-1]
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(payload) {
		return nil, errors.New("invalid payload encoding")
	}
	pkg := &testPackage{
		payload: string(payload),
	}
	var header []json.RawMessage
	err = json.Unmarshal(buf, &header)
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
		return 0, errors.New("write error simulation")
	}
	s, _ := w.buf.Write(b)
	return s, nil
}

func (w *testWriter) skip(c int) error {
	for i := range c {
		_, err := w.readPackage()
		if err != nil {
			return errors.Join(err, errors.New("failed"+ fmt.Sprint(i)))
		}
	}
	return nil
}

