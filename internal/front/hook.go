package front

import (
	"encoding/json"
	"github.com/doors-dev/doors/internal/door"
)


type Hook struct {
	Error    []*ErrorAction
	Scope    []*ScopeSet
	Indicate []*Indicate
	*door.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	a := []any{h.DoorId, h.HookId, h.Scope, h.Indicate, h.Error}
	return json.Marshal(a)
}
