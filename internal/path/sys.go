package path

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/doors-dev/doors/internal/resources"
	"github.com/mr-tron/base58"
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
	Door     uint64
	Hook     uint64
	Track    uint64
}

type UndoPath struct {
	Instance string
	Location Location
}

type resource [16]byte

type sync string

type Match struct {
	entity any
}

func (m Match) Hook() (HookMatch, bool) {
	e, ok := m.entity.(HookMatch)
	return e, ok
}

func (m Match) Resource() (hash [16]byte, ok bool) {
	e, ok := m.entity.(resource)
	return [16]byte(e), ok
}

func (m Match) Sync() (instanceID string, ok bool) {
	e, ok := m.entity.(sync)
	return string(e), ok
}

func (m Match) Undo() (UndoPath, bool) {
	e, ok := m.entity.(UndoPath)
	return e, ok
}

var hookRegexp = regexp.MustCompile(`^/h/([0-9a-zA-Z]+)/(\d+)/(\d+)(/[^/]+)?$`)
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
	path, ok := strings.CutPrefix(r.URL.Path, pm.Prefix())
	if !ok {
		return Match{}, false
	}
	matches := hookRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		instanceID := matches[1]
		doorID, err := strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return Match{}, false
		}
		hookID, err := strconv.ParseUint(matches[3], 10, 64)
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
				Door:     doorID,
				Hook:     hookID,
				Track:    track,
			},
		}, true
	}
	matches = resourceRegexp.FindStringSubmatch(path)
	if len(matches) != 0 {
		hashStr := matches[1]
		hash, err := base58.Decode(hashStr)
		if err != nil {
			return Match{}, false
		}
		if len(hash) != 16 {
			return Match{}, false
		}
		return Match{
			entity: resource(*(*[16]byte)(hash)),
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
		l, err := NewLocationFromEscapedPath(path)
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

func (pm PathMaker) Hook(instanceId string, doorId uint64, hookId uint64, name string) string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s/h/%s/%d/%d", pm.Prefix(), instanceId, doorId, hookId)
	if name != "" {
		builder.WriteByte('/')
		builder.WriteString(name)
	}
	return builder.String()
}

func (pm PathMaker) Resource(r *resources.Resource, name string) string {
	builder := &strings.Builder{}
	hash := r.Hash()
	hashStr := base58.Encode(hash[:])
	fmt.Fprintf(builder, "%s/r/%s", pm.Prefix(), hashStr)
	if name != "" {
		builder.WriteByte('.')
		builder.WriteString(name)
	}
	return builder.String()
}
