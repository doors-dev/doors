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

package front

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/actions"
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

func AttrsAppendSetter(attrs gox.Attrs, id uint64) {
	val := jsonAttrs([]any{id})
	attrs.Get("data-d0s").Set(val)
}

func AttrsAppendEmitter(attrs gox.Attrs, id uint64) {
	val := jsonAttrs([]any{id})
	attrs.Get("data-d0e").Set(val)
}

func AttrsSetParent(attrs gox.Attrs, parent uint64) {
	attrs.Get("data-d0p").Set(parentAttr(parent))
}

func AttrsSetDoor(attrs gox.Attrs, id uint64, container bool) {
	if container {
		attrs.Get("id").Set(fmt.Sprintf("d0r%d", id))
		return
	}
	attrs.Get("data-d0r").Set(fmt.Sprintf("%d", id))

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
	var arr jsonAttrs
	if prev, ok := prev.(jsonAttrs); ok {
		arr = prev
	}
	arr = append(arr, j...)
	return arr
}

type parentAttr uint64

var _ gox.Mutate = parentAttr(0)

func (p parentAttr) Mutate(_ string, prev any) any {
	if prev != nil {
		return prev
	}
	return p
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
	payload, err := actions.IntoPayload(j.value, false)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(common.NewJsonWriter(w))
	return enc.Encode(payload)
}
