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
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal"
	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/zeebo/xxh3"
)

type App interface {
	Logger() *slog.Logger
	Conf() *common.Conf
	ESProfile(profile string) api.BuildOptions
}

type Registry = *registry

func NewRegistry(app App) Registry {
	return &registry{
		app: app,
	}
}

type registry struct {
	initGuard  sync.Once
	app        App
	cache      sync.Map
	lookup     sync.Map
	mainScript *Resource
	mainStyle  *Resource
}

func (rs *registry) init() {
	opt := rs.app.ESProfile("")
	ScriptFS{
		FS:   internal.ClientSrc,
		Path: "index.ts",
		Name: "doors",
	}.Apply(&opt)
	FormatIIFE{
		GlobalName: "_d0r",
		Bundle:     true,
	}.Apply(&opt)
	opt.MangleProps = "_$"
	opt.Footer = map[string]string{
		"js": "_d0r = _d0r.default;",
	}
	content, err := build(&opt)
	if err != nil {
		rs.app.Logger().Error("esbuild error", "error", err)
		panic(errors.Join(errors.New("client JS build error"), err))
	}
	rs.mainScript = NewResource(content, "application/javascript", rs.defaultSettings())
	rs.lookup.Store(rs.mainScript.id, rs.mainScript)
	rs.mainStyle = NewResource(internal.ClientStyles, "text/css", rs.defaultSettings())
	rs.lookup.Store(rs.mainStyle.id, rs.mainStyle)
}

func (rs *registry) defaultSettings() resourceSettings {
	return resourceSettings{
		cacheControl: rs.app.Conf().ServerCacheControl,
		disableGzip:  rs.app.Conf().ServerDisableGzip,
	}
}

func (rs Registry) MainStyle() *Resource {
	rs.initGuard.Do(rs.init)
	return rs.mainStyle
}

func (rs Registry) MainScript() *Resource {
	rs.initGuard.Do(rs.init)
	return rs.mainScript
}

func (rs Registry) Serve(id string, w http.ResponseWriter, r *http.Request) {
	s, ok := rs.lookup.Load(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.(*Resource).Serve(w, r)
}

func (rs *registry) create(key [16]byte, content []byte, lookup bool, contentType string) *Resource {
	s := NewResource(content, contentType, rs.defaultSettings())
	existing, existed := rs.cache.LoadOrStore(key, s)
	if existed {
		s = existing.(*Resource)
	}
	if lookup {
		rs.lookup.Store(s.id, s)
	}
	return s
}

func (rs *registry) get(key [16]byte) *Resource {
	entry, ok := rs.cache.Load(key)
	if !ok {
		return nil
	}
	return entry.(*Resource)
}

func (rs Registry) Static(entry StaticEntry, contentType string) (*Resource, error) {
	h := xxh3.New()
	h.WriteString("resource")
	entry.entryID(h)
	key := h.Sum128().Bytes()
	res := rs.get(key)
	if res != nil {
		return res, nil
	}
	content, err := entry.Read()
	if err != nil {
		return nil, err
	}
	res = rs.create(key, content, true, contentType)
	return res, nil
}

func (rs Registry) Script(entry ScriptEntry, format ScriptFormat, profile string, mode ResourceMode) (*Resource, error) {
	var res *Resource
	var key [16]byte
	if mode != ModeNoCache {
		h := xxh3.New()
		h.WriteString("script")
		h.WriteString(profile)
		entry.entryID(h)
		format.formatID(h)
		key = h.Sum128().Bytes()
		res = rs.get(key)
	}
	if res != nil {
		return res, nil
	}
	var content []byte
	var err error
	if _, ok := format.(FormatRaw); ok {
		content, err = entry.Read()
	} else {
		opt := rs.app.ESProfile(profile)
		err := entry.Apply(&opt)
		if err != nil {
			return nil, err
		}
		format.Apply(&opt)
		content, err = build(&opt)
		if err != nil {
			rs.app.Logger().Error("esbuild error", "error", err)
		}
	}
	if err != nil {
		return nil, err
	}
	if mode != ModeNoCache {
		res = rs.create(key, content, mode == ModeHost, "application/javascript")
	} else {
		res = NewResource(content, "application/javascript", rs.defaultSettings())
	}
	return res, nil
}

func (rs Registry) Style(entry StyleEntry, minify bool, mode ResourceMode) (*Resource, error) {
	var res *Resource
	var key [16]byte
	if mode != ModeNoCache {
		h := xxh3.New()
		h.WriteString("style")
		entry.entryID(h)
		key = h.Sum128().Bytes()
		res = rs.get(key)
	}
	if res != nil {
		return res, nil
	}
	var content, err = entry.Read()
	if err == nil && minify {
		content, err = common.MinifyCSS(content)
	}
	if err != nil {
		return nil, err
	}
	if mode != ModeNoCache {
		res = rs.create(key, content, mode == ModeHost, "text/css")
	} else {
		res = NewResource(content, "text/css", rs.defaultSettings())
	}
	return res, nil
}

type ResourceMode int

const (
	ModeHost ResourceMode = iota
	ModeNoHost
	ModeNoCache
)
