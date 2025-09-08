// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package resources

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/mr-tron/base58"
	"github.com/zeebo/blake3"
)

func NewResource(content []byte, contentType string, s settings) *Resource {
	hash := blake3.Sum256(content)
	shortHash := hash[:16]
	return &Resource{
		hash:        *(*[16]byte)(shortHash),
		hashString:  base58.Encode(shortHash),
		settings:    s,
		content:     content,
		contentType: contentType,
	}
}

type InlineResource struct {
	Attrs    templ.Attributes
	resource *Resource
}

func (i *InlineResource) Resource() *Resource {
	return i.resource
}

func (i *InlineResource) Content() []byte {
	return i.resource.content
}

func (i *InlineResource) Serve(w http.ResponseWriter, r *http.Request) {
	i.resource.ServeCache(w, r, false)
}

type Resource struct {
	hashString  string
	settings    settings
	hash        [16]byte
	once        sync.Once
	content     []byte
	gzipped     []byte
	contentType string
}

func (s *Resource) HashString() string {
	return s.hashString
}

func (s *Resource) Content() []byte {
	return s.content
}

func (s *Resource) ServeCache(w http.ResponseWriter, r *http.Request, cache bool) {
	w.Header().Set("Content-Type", s.contentType)
	if cache {
		w.Header().Set("Cache-Control", s.settings.Conf().ServerCacheControl)
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if !s.settings.Conf().ServerDisableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		s.once.Do(func() {
			zipped, err := common.Zip(s.content)
			if err != nil {
				slog.Error("gzip error: " + err.Error())
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
