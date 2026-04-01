// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package resources

import (
	"errors"
	"github.com/doors-dev/doors/internal"
	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/zeebo/blake3"
	"net/http"
	"sync"
)

type BuildProfiles interface {
	Options(profile string) api.BuildOptions
}

type settings interface {
	Conf() *common.SystemConf
	BuildProfiles() BuildProfiles
}

func NewRegistry(s settings) *Registry {
	return &Registry{
		settings: s,
	}
}

type Registry struct {
	initGuard  sync.Once
	settings   settings
	cache      sync.Map
	lookup     sync.Map
	mainScript *Resource
	mainStyle  *Resource
}

func (rg *Registry) key32(b []byte) [32]byte {
	return *(*[32]byte)(b)
}

func (rg *Registry) init() {
	opt := rg.settings.BuildProfiles().Options("")
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
		panic(errors.Join(errors.New("Client js build error"), err))
	}
	rg.mainScript = NewResource(content, "application/javascript", rg.defaultSettings())
	rg.lookup.Store(rg.mainScript.id, rg.mainScript)
	rg.mainStyle = NewResource(internal.ClientStyles, "text/css", rg.defaultSettings())
	rg.lookup.Store(rg.mainStyle.id, rg.mainStyle)
}

func (rg *Registry) defaultSettings() resourceSettings {
	return resourceSettings{
		cacheControl: rg.settings.Conf().ServerCacheControl,
		disableGzip:  rg.settings.Conf().ServerDisableGzip,
	}
}

func (rg *Registry) MainStyle() *Resource {
	rg.initGuard.Do(rg.init)
	return rg.mainStyle
}

func (rg *Registry) MainScript() *Resource {
	rg.initGuard.Do(rg.init)
	return rg.mainScript
}

func (rg *Registry) Serve(id string, w http.ResponseWriter, r *http.Request) {
	s, ok := rg.lookup.Load(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.(*Resource).Serve(w, r)
}

func (r *Registry) create(key []byte, content []byte, lookup bool, contentType string) *Resource {
	s := NewResource(content, contentType, r.defaultSettings())
	existing, existed := r.cache.LoadOrStore(r.key32(key), s)
	if existed {
		s = existing.(*Resource)
	}
	if lookup {
		r.lookup.Store(s.id, s)
	}
	return s
}

func (r *Registry) get(key []byte) *Resource {
	entry, ok := r.cache.Load(r.key32(key))
	if !ok {
		return nil
	}
	return entry.(*Resource)
}

func (r *Registry) Static(entry StaticEntry, contentType string) (*Resource, error) {
	h := blake3.New()
	h.WriteString("resource")
	entry.entryID(h)
	key := h.Sum(nil)
	res := r.get(key)
	if res != nil {
		return res, nil
	}
	content, err := entry.Read()
	if err != nil {
		return nil, err
	}
	res = r.create(key, content, true, contentType)
	return res, nil
}

func (r *Registry) Script(entry ScriptEntry, format ScriptFormat, profile string, mode ResourceMode) (*Resource, error) {
	var res *Resource
	var key []byte
	if mode != ModeNoCache {
		h := blake3.New()
		h.WriteString("script")
		h.WriteString(profile)
		entry.entryID(h)
		format.formatID(h)
		key = h.Sum(nil)
		res = r.get(key)
	}
	if res != nil {
		return res, nil
	}
	var content []byte
	var err error
	if _, ok := format.(FormatRaw); ok {
		content, err = entry.Read()
	} else {
		opt := r.settings.BuildProfiles().Options(profile)
		err := entry.Apply(&opt)
		if err != nil {
			return nil, err
		}
		format.Apply(&opt)
		content, err = build(&opt)
	}
	if err != nil {
		return nil, err
	}
	if key != nil {
		res = r.create(key, content, mode == ModeHost, "application/javascript")
	} else {
		res = NewResource(content, "application/javascript", r.defaultSettings())
	}
	return res, nil
}

func (r *Registry) Style(entry StyleEntry, minify bool, mode ResourceMode) (*Resource, error) {
	var res *Resource
	var key []byte
	if mode != ModeNoCache {
		h := blake3.New()
		h.WriteString("style")
		entry.entryID(h)
		key = h.Sum(nil)
		res = r.get(key)
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
	if key != nil {
		res = r.create(key, content, mode == ModeHost, "text/css")
	} else {
		res = NewResource(content, "text/css", r.defaultSettings())
	}
	return res, nil
}

type ResourceMode int

const (
	ModeHost ResourceMode = iota
	ModeNoHost
	ModeNoCache
)
