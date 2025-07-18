package front

import (
	"encoding/json"
	"github.com/doors-dev/doors/internal/node"
)

var null = []byte("null")

type Hook struct {
	Mark      string
	Scope     []*ScopeSet
	Indicate []*Indicate
	*node.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	/*
		scopes := make([]ScopeSet, len(h.Scope))
		for i, scope := range h.Scope {
			scopes[i] = sco
		} */
	a := []any{h.NodeId, h.HookId, h.Scope, h.Indicate, h.Mark}
	return json.Marshal(a)
}
