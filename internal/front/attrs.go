package front

import (
	"fmt"

	"github.com/doors-dev/gox"
)

func AttrsAppendCapture(attrs gox.Attrs, capture Capture, hook Hook) {
	attrs.Get("data-d00r-capture").AppendObject([]any{capture.Listen(), capture.Name(), capture, hook})
}

func AttrsSetHook(attrs gox.Attrs, name string, hook Hook) {
	attrs.Get(fmt.Sprintf("data-d00r-hook:%s", name)).SetObject(hook)
}

func AttrsSetData(attrs gox.Attrs, name string, data any) {
	attrs.Get(fmt.Sprintf("data-d00r-data:%s", name)).SetObject(data)
}

func AttrsAppendDyna(attrs gox.Attrs, id uint64, name string) {
	attrs.Get("data-d00r-dyna").AppendObject([]any{id, name})
}

