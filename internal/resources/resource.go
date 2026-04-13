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
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/zeebo/blake3"
)

type resourceSettings struct {
	cacheControl string
	disableGzip  bool
}

func NewResource(content []byte, contentType string, s resourceSettings) *Resource {
	hash := blake3.Sum256(content)
	return &Resource{
		id:          common.EncodeId(hash[:]),
		settings:    s,
		content:     content,
		contentType: contentType,
	}
}

type Resource struct {
	id          string
	settings    resourceSettings
	once        sync.Once
	content     []byte
	gzipped     []byte
	contentType string
}

func (s *Resource) ID() string {
	return s.id
}

/*
func (s *Resource) HashString() string {
	return s.hashString
} */

func (s *Resource) Content() []byte {
	return s.content
}

func (s *Resource) ServeCache(w http.ResponseWriter, r *http.Request, cache bool) {
	if s.contentType != "" {
		w.Header().Set("Content-Type", s.contentType)
	}
	if cache {
		w.Header().Set("Cache-Control", s.settings.cacheControl)
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if !s.settings.disableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		s.once.Do(func() {
			zipped, err := common.Zip(s.content)
			if err != nil {
				slog.Error("gzip error", "error", err)
			}
			s.gzipped = zipped
		})
		if s.gzipped != nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(s.gzipped)
			return
		}
	}
	w.Write(s.content)
}
func (s *Resource) Serve(w http.ResponseWriter, r *http.Request) {
	s.ServeCache(w, r, true)
}
