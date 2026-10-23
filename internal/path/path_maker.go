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

package path

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/resources"
)

func NewPathMaker(prefix string, serverID string, serverIDCookieName string) PathMaker {
	return PathMaker{
		sessionCookie:      prefix + serverID,
		serverID:           serverID,
		serverIDCookieName: serverIDCookieName,
	}
}

type PathMaker struct {
	serverID           string
	sessionCookie      string
	serverIDCookieName string
}

type HookMatch struct {
	Instance string
	Hook     uint64
	Track    uint64
	Prefix   string
	Path     string
	RawPath  string
}

type Rewrite struct {
	Path    string
	RawPath string
	Context func(context.Context) context.Context
}

type UndoPath struct {
	Instance string
	Location Location
}

type resourceEnt string

type syncEnt string

type tabStateEnt string

type Match struct {
	entity any
}

func (m Match) Hook() (HookMatch, bool) {
	e, ok := m.entity.(HookMatch)
	return e, ok
}

func (m Match) Resource() (id string, ok bool) {
	e, ok := m.entity.(resourceEnt)
	return string(e), ok
}

func (m Match) Sync() (instanceID string, ok bool) {
	e, ok := m.entity.(syncEnt)
	return string(e), ok
}

func (m Match) Undo() (UndoPath, bool) {
	e, ok := m.entity.(UndoPath)
	return e, ok
}

func (m Match) TabState() (instanceID string, ok bool) {
	e, ok := m.entity.(tabStateEnt)
	return string(e), ok
}

var hookRegexp = regexp.MustCompile(`^/h/([0-9a-zA-Z]+)/(\d+)(\?.*|/.*)?$`)
var resourceRegexp = regexp.MustCompile(`^/r/([0-9a-zA-Z]+)(\.[^/]+)?$`)
var syncPath = regexp.MustCompile(`^/s/([0-9a-zA-Z]+)(\?.*)?$`)
var undoPath = regexp.MustCompile(`^/u/([0-9a-zA-Z]+)(/.*)$`)
var tabStatePath = regexp.MustCompile(`^/t/([0-9a-zA-Z]+)(\?.*)?$`)
var portablePath = regexp.MustCompile(`^/w/([0-9a-zA-Z]+)(/[^?]*)?(\?.*)?$`)

func (pm PathMaker) SetServerIDCookie() bool {
	return pm.serverIDCookieName != ""
}

func (pm PathMaker) ServerIDCookieName() string {
	return pm.serverIDCookieName
}

func (pm PathMaker) ID() string {
	return pm.serverID
}

func (pm PathMaker) SessionCookie() string {
	return pm.sessionCookie
}

const systemPrefix = "/~/"

func (pm PathMaker) Prefix() string {
	return systemPrefix + pm.serverID
}

func (pm PathMaker) IsSystem(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, systemPrefix)
}

func (pm PathMaker) Rewrite(r *http.Request) (token string, tail string, ok bool) {
	tilde, ok := strings.CutPrefix(r.URL.RequestURI(), systemPrefix)
	if !ok {
		return "", "", false
	}
	index := strings.IndexByte(tilde, '/')
	if index < 0 {
		return "", "", false
	}
	matches := portablePath.FindStringSubmatch(tilde[index:])
	if len(matches) == 0 {
		return "", "", false
	}
	if matches[2] != "" {
		if _, err := url.ParseRequestURI(matches[2]); err != nil {
			return "", "", false
		}
	}
	return matches[1], matches[2], true
}

func (pm PathMaker) Match(r *http.Request) (Match, bool) {
	uri := r.URL.RequestURI()
	tilde, ok := strings.CutPrefix(uri, systemPrefix)
	if !ok {
		return Match{}, false
	}
	index := strings.IndexByte(tilde, '/')
	if index < 0 {
		return Match{}, false
	}
	path := tilde[index:]
	matches := hookRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		hookID, err := strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return Match{}, false
		}
		track := uint64(0)
		trackStr := r.Header.Get("D0-Track")
		if trackStr != "" {
			track, err = strconv.ParseUint(trackStr, 10, 64)
			if err != nil {
				return Match{}, false
			}
		}
		tail, _, _ := strings.Cut(matches[3], "?")
		if tail == "" {
			tail = "/"
		}
		u, err := url.ParseRequestURI(tail)
		if err != nil {
			return Match{}, false
		}
		prefix, ok := r.Context().Value(common.KeyHookPrefix).(string)
		if !ok {
			prefix = uri[:len(uri)-len(matches[3])]
		}
		return Match{
			entity: HookMatch{
				Instance: instanceID,
				Hook:     hookID,
				Track:    track,
				Prefix:   prefix,
				Path:     u.Path,
				RawPath:  u.RawPath,
			},
		}, true
	}
	matches = resourceRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		id := matches[1]
		return Match{
			entity: resourceEnt(id),
		}, true
	}
	matches = syncPath.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		return Match{
			entity: syncEnt(instanceID),
		}, true
	}
	matches = undoPath.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		path := matches[2]
		l := NewLocationFromEscapedURI(path)
		return Match{
			entity: UndoPath{
				Instance: instanceID,
				Location: l,
			},
		}, true
	}
	matches = tabStatePath.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		return Match{
			entity: tabStateEnt(instanceID),
		}, true
	}
	return Match{}, false
}

func (pm PathMaker) Hook(instanceId string, hookId uint64, name string) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s/h/%s/%d", pm.Prefix(), instanceId, hookId)
	if name != "" {
		builder.WriteByte('/')
		builder.WriteString(url.PathEscape(name))
	}
	return builder.String()
}

func (pm PathMaker) HookTail(instanceId string, hookId uint64, tail string) (path string, rawPath string) {
	u, err := url.ParseRequestURI(pm.Hook(instanceId, hookId, "") + tail)
	if err != nil {
		return pm.Hook(instanceId, hookId, ""), ""
	}
	return u.Path, u.RawPath
}

func (pm PathMaker) Portable(token string, name string) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s/w/%s", pm.Prefix(), token)
	if name != "" {
		builder.WriteByte('/')
		builder.WriteString(url.PathEscape(name))
	}
	return builder.String()
}

func (pm PathMaker) Resource(r *resources.Resource, name string) string {
	builder := &strings.Builder{}
	id := r.ID()
	fmt.Fprintf(builder, "%s/r/%s", pm.Prefix(), id)
	if name != "" {
		builder.WriteByte('.')
		builder.WriteString(url.PathEscape(name))
	}
	return builder.String()
}
