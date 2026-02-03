package front

import (
	"fmt"

	"github.com/doors-dev/gox"
)

func AttrsAppendCapture(attrs gox.Attrs, capture Capture, hook Hook) {
	attrs.Get("data-d0c").AppendObject([]any{capture.Listen(), capture.Name(), capture, hook})
}

func AttrsSetHook(attrs gox.Attrs, name string, hook Hook) {
	attrs.Get(fmt.Sprintf("data-d0h-%s", name)).SetObject(hook)
}

func AttrsSetData(attrs gox.Attrs, name string, data any) {
	attrs.Get(fmt.Sprintf("data-d0d-%s", name)).SetObject(data)
}

func AttrsAppendDyna(attrs gox.Attrs, id uint64, name string) {
	attrs.Get("data-d0y").AppendObject([]any{id, name})
}

func AttrsSetActive(attrs gox.Attrs, active []any) {
	attrs.Get("data-d0a").SetObject(active)
}
