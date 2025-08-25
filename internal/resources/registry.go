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
	"bytes"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"sync"

	"github.com/doors-dev/doors/internal"
	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/zeebo/blake3"
)

type BuildProfiles interface {
	Options(profile string) api.BuildOptions
}

func NewRegistry() *Registry {
	return &Registry{
		Gzip:     true,
		Profiles: BaseProfile{},
	}
}

type Registry struct {
	Gzip       bool
	Profiles   BuildProfiles
	cache      sync.Map
	lookup     sync.Map
	mainScript *Resource
	mainStyle  *Resource
	init       sync.Once
}

func (rg *Registry) key(b []byte) [32]byte {
	return *(*[32]byte)(b)
}

func (rg *Registry) initMain() {
	rg.init.Do(func() {
		profile := rg.Profiles.Options("")
		profile.Format = api.FormatIIFE
		profile.Footer = map[string]string{
			"js": "_d00r = _d00r.default;",
		}
		profile.Bundle = true
		profile.GlobalName = "_d00r"
		scriptContent, err := BuildFS(internal.ClientSrc, "index.ts", profile)
		if err != nil {
			panic(errors.Join(errors.New("Client js build error"), err))
		}
		rg.mainScript = NewResource(scriptContent, "application/javascript", rg.Gzip)
		rg.mainStyle = NewResource(internal.ClientStyles, "text/css", rg.Gzip)
	})
}

func (rg *Registry) MainStyle() *Resource {
	rg.initMain()
	return rg.mainStyle
}

func (rg *Registry) MainScript() *Resource {
	rg.initMain()
	return rg.mainScript
}

func (rg *Registry) Serve(hash []byte, w http.ResponseWriter, r *http.Request) {
	if len(hash) != 32 {
		w.WriteHeader(400)
		return
	}
	s, ok := rg.lookup.Load(rg.key(hash))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.(*Resource).Serve(w, r)
}

func (r *Registry) create(key []byte, content []byte, lookup bool, contentType string) *Resource {
	s := NewResource(content, contentType, r.Gzip)
	existing, existed := r.cache.LoadOrStore(r.key(key), s)
	if existed {
		s = existing.(*Resource)
	}
	if lookup {
		r.lookup.Store(s.hash, s)
	}
	return s
}

func (r *Registry) get(key []byte) *Resource {
	entry, ok := r.cache.Load(r.key(key))
	if !ok {
		return nil
	}
	return entry.(*Resource)
}

func (r *Registry) StyleBytes(content []byte) (*Resource, error) {
	h := blake3.New()
	h.WriteString("style_bytes")
	h.Write(content)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	min, err := common.MinifyCSS(content)
	if err != nil {
		return nil, err
	}
	return r.create(key, min, true, "text/css"), nil
}
func (r *Registry) Style(path string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("style")
	h.WriteString(path)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	min, err := common.MinifyCSS(content)
	if err != nil {
		return nil, err
	}
	return r.create(key, min, true, "text/css"), nil
}

type InlineMode int

const (
	InlineModeHost InlineMode = iota
	InlineModeLocal
	InlineModeNoCache
)

func (r *Registry) InlineStyle(data []byte, mode InlineMode) (*InlineResource, error) {
	element, err := HTMLParseElement("style", data)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, nil
	}
	h := blake3.New()
	h.WriteString("inline_style")
	h.Write(data)
	key := h.Sum(nil)
	var s *Resource
	if mode != InlineModeNoCache {
		s = r.get(key)
	}
	if s == nil {
		min, err := common.MinifyCSS(element.Content)
		if err != nil {
			return nil, err
		}
		if mode == InlineModeNoCache {
			s = NewResource(min, "text/css", r.Gzip)
		} else {
			s = r.create(key, min, mode == InlineModeHost, "text/css")
		}
	}
	return &InlineResource{
		Attrs:    element.Attrs,
		resource: s,
	}, nil

}

func (r *Registry) InlineScript(data []byte, mode InlineMode) (*InlineResource, error) {
	element, err := HTMLParseElement("script", data)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, nil
	}
	attr, ok := element.Attrs["type"]
	ts := false
	if ok {
		t := attr.(string)
		if t == "module" {
			return nil, errors.New("Module scripts are disallowed. Check the docs")
		}
		if t == "text/typescript" || t == "application/typescript" {
			ts = true
			delete(element.Attrs, "type")
		}
	}
	h := blake3.New()
	if ts {
		h.WriteString("inline_ts")
	} else {
		h.WriteString("inline_js")
	}
	h.Write(data)
	key := h.Sum(nil)
	var s *Resource
	if mode != InlineModeNoCache {
		s = r.get(key)
	}
	if s == nil {
		buf := &bytes.Buffer{}
		buf.WriteString("_d00r(document.currentScript, async ($d) => {\n")
		buf.Write(element.Content)
		buf.WriteString("\n})")
		var content []byte
		if ts {
			content, err = TransformBytesTS(buf.Bytes(), r.Profiles.Options(""))
			if err != nil {
				return nil, err
			}
		} else {
			content, err = TransformBytes(buf.Bytes(), r.Profiles.Options(""))
			if err != nil {
				return nil, err
			}
		}
		if mode == InlineModeNoCache {
			s = NewResource(content, "application/json", r.Gzip)
		} else {
			s = r.create(key, content, mode == InlineModeHost, "application/javascript")
		}
	}
	return &InlineResource{
		Attrs:    element.Attrs,
		resource: s,
	}, nil

}

func (r *Registry) ModuleBundle(entry string, profile string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("bundle")
	h.WriteString(entry)
	h.WriteString(profile)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	countent, err := Bundle(entry, r.Profiles.Options(profile))
	if err != nil {
		return nil, err
	}
	return r.create(key, countent, true, "application/javascript"), nil

}

func (r *Registry) ModuleBundleFS(cacheKey string, fs fs.FS, entry string, profile string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("bundle_fs")
	h.WriteString(cacheKey)
	h.WriteString(entry)
	h.WriteString(profile)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	countent, err := BundleFS(fs, entry, r.Profiles.Options(profile))
	if err != nil {
		return nil, err
	}
	return r.create(key, countent, true, "application/javascript"), nil
}

func (r *Registry) Module(path string, profile string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("module")
	h.WriteString(path)
	h.WriteString(profile)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	countent, err := Transform(path, r.Profiles.Options(profile))
	if err != nil {
		return nil, err
	}
	return r.create(key, countent, true, "application/javascript"), nil

}

func (r *Registry) ModuleBytes(content []byte, profile string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("bytes")
	h.Write(content)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	countent, err := TransformBytes(content, r.Profiles.Options(profile))
	if err != nil {
		return nil, err
	}
	return r.create(key, countent, true, "application/javascript"), nil
}

func (r *Registry) ModuleBytesTS(content []byte, profile string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("bytes_ts")
	h.Write(content)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	countent, err := TransformBytesTS(content, r.Profiles.Options(profile))
	if err != nil {
		return nil, err
	}
	return r.create(key, countent, true, "application/javascript"), nil
}

func (r *Registry) ModuleRaw(path string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("raw")
	h.WriteString(path)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return r.create(key, content, true, "application/javascript"), nil
}

func (r *Registry) ModuleRawBytes(content []byte) (*Resource, error) {
	h := blake3.New()
	h.WriteString("raw_bytes")
	h.Write(content)
	key := h.Sum(nil)
	s := r.get(key)
	if s != nil {
		return s, nil
	}
	return r.create(key, content, true, "application/javascript"), nil
}
