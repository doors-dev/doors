package front

import (
	"encoding/json"
	"github.com/doors-dev/doors/internal/node"
)


type Hook struct {
	Error    []*ErrorAction
	Scope    []*ScopeSet
	Indicate []*Indicate
	*node.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	a := []any{h.NodeId, h.HookId, h.Scope, h.Indicate, h.Error}
	return json.Marshal(a)
}
