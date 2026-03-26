// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
	"time"
	"unsafe"

	"github.com/doors-dev/gox"
	"github.com/mr-tron/base58"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

func AttrsToMap(a gox.Attrs) map[string]string {
	attrs := make(map[string]string)
	b := &bytes.Buffer{}
	for _, attr := range a.List() {
		if !attr.IsSet() {
			continue
		}
		if err := attr.OutputName(b); err != nil {
			slog.Error("Can't write attr name", "err", err)
			continue
		}
		name := b.String()
		b.Reset()
		if err := attr.OutputValue(b); err != nil {
			slog.Error("Can't write attr value", "err", err)
			continue
		}
		value := b.String()
		b.Reset()
		attrs[name] = value
	}
	return attrs
}

var bytesNull = []byte("null")

/*
func MarshalJSON(value any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(value)
	if err != nil {
		return bytesNull, err
	}
	b := StripN(buf.Bytes())
	return b, nil
} */

func Ts() {
	fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
}
func AsString(buf *[]byte) string {
	return *(*string)(unsafe.Pointer(buf))
}

func AsBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func RandId() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("failed to generate random bytes: %v", err)
	}
	return EncodeId(randomBytes)
}

const idLen = 22

func EncodeId(b []byte) string {
	s := base58.Encode(b)
	if len(s) == 0 {
		return s
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		if len(s)-i <= idLen {
			rest := s[i:]
			if len(rest) > idLen {
				rest = rest[:idLen]
			}
			out := make([]byte, len(rest))
			copy(out, rest)
			out[0] = digitToLetter[out[0]-'0']
			return string(out)
		}
		i++
	}
	rest := s[i:]
	if len(rest) > idLen {
		rest = rest[:idLen]
	}
	return rest
}

var digitToLetter = [10]byte{
	'z', // '0'
	'o', // '1'
	't', // '2'
	'r', // '3'
	'f', // '4'
	'i', // '5'
	's', // '6'
	'v', // '7'
	'e', // '8'
	'n', // '9'
}

func Zip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(input)
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MinifyCSS(input []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	return m.Bytes("text/css", input)
}

func Hash(input []byte) string {
	hash := crypto.SHA3_224.New()
	hash.Write(input)
	return base58.Encode(hash.Sum(nil)[0:12])
}

func Catch(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	f()
	return
}

func CatchValue[V any](f func() V) (value V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	value = f()
	return
}
