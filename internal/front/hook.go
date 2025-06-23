package front

import (
	"encoding/json"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
)

var null = []byte("null")

type SelectorMode string



const (
	SelectModeTarget SelectorMode = "target"
	SelectModeQuery = "query"
	SelectModeParentQuery = "parent_query"
)

type Selector struct {
	mode  SelectorMode
	query string
}

func SelectTarget() *Selector {
	return &Selector{
		mode: SelectModeTarget,
	}
}
func SelectQuery(query string) *Selector {
	return &Selector{
		mode:  SelectModeQuery,
		query: query,
	}
}
func SelectParentQuery(query string) *Selector {
	return &Selector{
		mode:  SelectModeParentQuery,
		query: query,
	}
}

func (s *Selector) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{string(s.mode), s.query})
}

type Indicate struct {
	selector *Selector
	kind     string
	param1   string
	param2   string
}

func IndicateAttr(s *Selector, name string, value string) Indicate {
	return Indicate{
		selector: s,
		kind:     "attr",
		param1:   name,
		param2:   value,
	}
}

func IndicateClass(s *Selector, class string) Indicate {
	return Indicate{
		selector: s,
		kind:     "class",
		param1:   class,
	}
}

func IndicateClassRemove(s *Selector, class string) Indicate {
	return Indicate{
		selector: s,
		kind:     "remove_class",
		param1:   class,
	}
}

func IndicateContent(s *Selector, content string) Indicate {
	return Indicate{
		selector: s,
		kind:     "content",
		param1:   content,
	}
}

func (i *Indicate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{i.selector, i.kind, i.param1, i.param2})
}

type Hook struct {
	Mark      string
	Mode      HookMode
	Indicate []Indicate
	*node.HookEntry
}

func (h *Hook) MarshalJSON() ([]byte, error) {
	if common.IsNill(h.Mode) {
		h.Mode = Default()
	}
	a := []any{h.NodeId, h.HookId, h.Mode, h.Indicate, h.Mark}
	return json.Marshal(a)
}
