package action

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"

	"github.com/doors-dev/doors/internal/common"
)

func NewNone() Payload {
	return Payload{}
}
func NewTextBytes(text []byte) Payload {
	return Payload{entity: TextBytes(text)}
}

func NewText(v string) Payload {
	return Payload{entity: Text(v)}
}

func NewTextGZ(v []byte) Payload {
	return Payload{entity: TextGZ(v)}
}

func NewJSON(v []byte) Payload {
	return Payload{entity: JSON(v)}
}

func NewJSONGZ(v []byte) Payload {
	return Payload{entity: JSONGZ(v)}
}

func NewBinary(v []byte) Payload {
	return Payload{entity: Binary(v)}
}

func NewBinaryGZ(v []byte) Payload {
	return Payload{entity: BinaryGZ(v)}
}

func NewEmptyPayload() Payload {
	return Payload{}
}

type TextBytes []byte

func (t TextBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

type Text string

type TextGZ []byte

type JSON []byte

func (j JSON) MarshalJSON() ([]byte, error) {
	return j, nil
}

type JSONGZ []byte

type Binary []byte

type BinaryGZ []byte

type Payload struct {
	entity any
}

func (p Payload) IsNone() bool {
	return p.entity == nil
}

func (p Payload) Type() PayloadType {
	if p.entity == nil {
		return PayloadNone
	}
	switch p.entity.(type) {
	case Text, TextBytes:
		return PayloadText
	case TextGZ:
		return PayloadTextGZ
	case JSON:
		return PayloadJSON
	case JSONGZ:
		return PayloadJSONGZ
	case Binary:
		return PayloadBinary
	case BinaryGZ:
		return PayloadBinaryGZ
	default:
		panic("unknown payload type")
	}
}

func (p Payload) Len() int {
	if p.entity == nil {
		return 0
	}
	switch v := p.entity.(type) {
	case Text:
		return len(v)
	case TextBytes:
		return len(v)
	case TextGZ:
		return len(v)
	case JSON:
		return len(v)
	case JSONGZ:
		return len(v)
	case Binary:
		return len(v)
	case BinaryGZ:
		return len(v)
	default:
		panic("unknown payload type")
	}
}

func (p Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{p.Type(), p.entity})
}

func (p Payload) Output(w io.Writer) error {
	if p.entity == nil {
		return nil
	}
	switch v := p.entity.(type) {
	case Text:
		_, err := io.WriteString(w, string(v))
		return err
	case TextBytes:
		_, err := w.Write(v)
		return err
	case TextGZ:
		_, err := w.Write(v)
		return err
	case JSON:
		_, err := w.Write(v)
		return err
	case JSONGZ:
		_, err := w.Write(v)
		return err
	case Binary:
		_, err := w.Write(v)
		return err
	case BinaryGZ:
		_, err := w.Write(v)
		return err
	default:
		panic("unknown payload type")
	}
}

type PayloadType int

const (
	PayloadNone     PayloadType = 0x00
	PayloadBinary   PayloadType = 0x01
	PayloadJSON     PayloadType = 0x02
	PayloadText     PayloadType = 0x03
	PayloadBinaryGZ PayloadType = 0x11
	PayloadJSONGZ   PayloadType = 0x12
	PayloadTextGZ   PayloadType = 0x13
)

func IntoPayload(v any, gz bool) (Payload, error) {
	if bytes, ok := v.([]byte); ok {
		if bytes == nil {
			bytes = make([]byte, 0)
		}
		return NewBinary(bytes), nil
	}
	buf := &bytes.Buffer{}
	var w io.Writer = buf
	var wgz *gzip.Writer
	if gz {
		wgz = gzip.NewWriter(buf)
		w = wgz
	}
	var t PayloadType
	if str, ok := v.(string); ok {
		if gz {
			io.WriteString(w, str)
			t = PayloadTextGZ
		} else {
			t = PayloadText
		}
	} else {
		encoder := json.NewEncoder(common.NewJsonWriter(w))
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(v); err != nil {
			return Payload{}, err
		}
		if gz {
			t = PayloadJSONGZ
		} else {
			t = PayloadJSON
		}
	}
	if gz {
		if err := wgz.Close(); err != nil {
			return Payload{}, err
		}
	}
	switch t {
	case PayloadJSONGZ:
		return NewJSONGZ(buf.Bytes()), nil
	case PayloadJSON:
		return NewJSON(buf.Bytes()), nil
	case PayloadText:
		return NewText(v.(string)), nil
	case PayloadTextGZ:
		return NewTextGZ(buf.Bytes()), nil
	default:
		panic("unexpected payload type")
	}

}
