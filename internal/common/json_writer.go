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
