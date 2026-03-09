package front

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

func AttrsAppendCapture(attrs gox.Attrs, capture Capture, hook Hook) {
	val := jsonAttrs([]any{[]any{capture.Listen(), capture.Name(), capture, hook}})
	attrs.Get("data-d0c").Set(val)
}

func AttrsSetHook(attrs gox.Attrs, name string, hook Hook) {
	attrs.Get(fmt.Sprintf("data-d0h-%s", name)).Set(jsonAttr{hook})
}

func AttrsSetData(attrs gox.Attrs, name string, data any) {
	attrs.Get(fmt.Sprintf("data-d0d-%s", name)).Set(payloadAttr{data})
}

func AttrsAppendDyn(attrs gox.Attrs, id uint64, name string) {
	val := jsonAttrs([]any{[]any{id, name}})
	attrs.Get("data-d0y").Set(val)
}

func AttrsSetActive(attrs gox.Attrs, active []any) {
	val := jsonAttr{active}
	attrs.Get("data-d0a").Set(val)
}

type jsonAttrs []any

func (j jsonAttrs) Output(w io.Writer) error {
	enc := json.NewEncoder(common.NewJsonWriter(w))
	enc.SetEscapeHTML(false)
	return enc.Encode(j)
}

var _ gox.Output = (jsonAttrs)(nil)
var _ gox.Mutate = (jsonAttrs)(nil)

func (j jsonAttrs) Mutate(name string, prev any) any {
	if !strings.HasPrefix(name, "data-d0") {
		slog.Error("Unexpected attribute name for system attribute", "name", name)
		return prev
	}
	var arr jsonAttrs
	if prev, ok := prev.(jsonAttrs); ok {
		arr = prev
	}
	arr = append(arr, j...)
	return arr
}

type jsonAttr struct {
	value any
}

func (j jsonAttr) Output(w io.Writer) error {
	enc := json.NewEncoder(common.NewJsonWriter(w))
	enc.SetEscapeHTML(false)
	return enc.Encode(j.value)
}

var _ gox.Output = jsonAttr{}

type payloadAttr struct {
	value any
}

func (j payloadAttr) Output(w io.Writer) error {
	payload, err := action.IntoPayload(j.value, false)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(common.NewJsonWriter(w))
	return enc.Encode(payload)
}
