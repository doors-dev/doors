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
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/doors-dev/doors/internal/resources"
)

func NewPathMaker(serverID string) PathMaker {
	if serverID == "" {
		serverID = "0"
	} else if strings.Contains(serverID, "/") {
		panic("server id can't contain \"/\"")
	}
	return PathMaker{
		serverID: serverID,
	}
}

type PathMaker struct {
	serverID string
}

type HookMatch struct {
	Instance string
	Hook     uint64
	Track    uint64
}

type UndoPath struct {
	Instance string
	Location Location
}

type resource string

type sync string

type Match struct {
	entity any
}

func (m Match) Hook() (HookMatch, bool) {
	e, ok := m.entity.(HookMatch)
	return e, ok
}

func (m Match) Resource() (id string, ok bool) {
	e, ok := m.entity.(resource)
	return string(e), ok
}

func (m Match) Sync() (instanceID string, ok bool) {
	e, ok := m.entity.(sync)
	return string(e), ok
}

func (m Match) Undo() (UndoPath, bool) {
	e, ok := m.entity.(UndoPath)
	return e, ok
}

var hookRegexp = regexp.MustCompile(`^/h/([0-9a-zA-Z]+)/(\d+)(\?.*|/.*)?$`)
var resourceRegexp = regexp.MustCompile(`^/r/([0-9a-zA-Z]+)(\.[^/]+)?$`)
var syncPath = regexp.MustCompile(`^/s/([0-9a-zA-Z]+)(/)?$`)
var undoPath = regexp.MustCompile(`^/u/([0-9a-zA-Z]+)(/.*)$`)

func (pm PathMaker) ID() string {
	return pm.serverID
}

func (pm PathMaker) Prefix() string {
	return "/~/" + pm.serverID
}

func (pm PathMaker) Match(r *http.Request) (Match, bool) {
	path, ok := strings.CutPrefix(r.URL.RequestURI(), pm.Prefix())
	if !ok {
		return Match{}, false
	}
	matches := hookRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		hookID, err := strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return Match{}, false
		}
		track := uint64(0)
		trackStr := r.URL.Query().Get("t")
		if trackStr != "" {
			track, err = strconv.ParseUint(trackStr, 10, 64)
			if err != nil {
				return Match{}, false
			}
		}
		return Match{
			entity: HookMatch{
				Instance: instanceID,
				Hook:     hookID,
				Track:    track,
			},
		}, true
	}
	matches = resourceRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		id := matches[1]
		return Match{
			entity: resource(id),
		}, true
	}
	matches = syncPath.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		return Match{
			entity: sync(instanceID),
		}, true
	}
	matches = undoPath.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		path := matches[2]
		l, err := NewLocationFromEscapedURI(path)
		if err != nil {
			return Match{}, false
		}
		return Match{
			entity: UndoPath{
				Instance: instanceID,
				Location: l,
			},
		}, true
	}
	return Match{}, false
}

func (pm PathMaker) Hook(instanceId string, hookId uint64, name string) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s/h/%s/%d", pm.Prefix(), instanceId, hookId)
	if name != "" {
		builder.WriteByte('/')
		builder.WriteString(name)
	}
	return builder.String()
}

func (pm PathMaker) Resource(r *resources.Resource, name string) string {
	builder := &strings.Builder{}
	id := r.ID()
	fmt.Fprintf(builder, "%s/r/%s", pm.Prefix(), id)
	if name != "" {
		builder.WriteByte('.')
		builder.WriteString(name)
	}
	return builder.String()
}
