// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"log"
	"log/slog"
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
			slog.Error("failed to write attribute name", "error", err)
			continue
		}
		name := b.String()
		b.Reset()
		if err := attr.OutputValue(b); err != nil {
			slog.Error("failed to write attribute value", "error", err)
			continue
		}
		value := b.String()
		b.Reset()
		attrs[name] = value
	}
	return attrs
}

func AsString(buf *[]byte) string {
	return *(*string)(unsafe.Pointer(buf))
}

func AsBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func RandId() string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatalf("failed to generate random bytes: %v", err)
	}
	return EncodeId(randomBytes)
}

const digitToLetter = "zotrfisven"

func EncodeId(b []byte) string {
	s := base58.Encode(b)
	if len(s) == 0 {
		return ""
	}
	c := s[0]
	if c < '0' || c > '9' {
		return s
	}
	return digitToLetter[c-'0':c-'0'+1] + s[1:]
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
