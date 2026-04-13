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

package resources

import (
	"io/fs"
	"os"

	"github.com/doors-dev/doors/internal/common"
)

type StyleEntry interface {
	Read() ([]byte, error)
	entryID(h idWriter)
}

type StyleFS struct {
	FS   fs.FS
	Path string
	Name string
}

func (e StyleFS) Read() ([]byte, error) {
	return fs.ReadFile(e.FS, e.Path)
}

func (e StyleFS) entryID(w idWriter) {
	w.WriteString("fs")
	w.WriteString(e.Path)
	w.WriteString(e.Name)
}

type StylePath struct {
	Path string
}

func (e StylePath) Read() ([]byte, error) {
	return os.ReadFile(e.Path)
}

func (e StylePath) entryID(w idWriter) {
	w.WriteString("path")
	w.WriteString(e.Path)
}

type StyleBytes struct {
	Content []byte
}

func (e StyleBytes) Read() ([]byte, error) {
	return e.Content, nil
}

func (e StyleBytes) entryID(w idWriter) {
	w.WriteString("content")
	w.Write(e.Content)
}

type StyleString struct {
	Content string
}

func (e StyleString) Read() ([]byte, error) {
	return common.AsBytes(e.Content), nil
}

func (e StyleString) entryID(w idWriter) {
	w.WriteString("content")
	w.WriteString(e.Content)
}
