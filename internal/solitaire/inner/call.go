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

package inner

import (
	"encoding/json"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/front/action"
)

type Call struct {
	Call     action.Call
	Params   action.CallParams
	reported atomic.Bool
}

func (p *Call) Written() {
	if !p.Params.Optimistic {
		return
	}
	p.Result([]byte("null"), nil)
}

func (c *Call) Action() (action.Action, bool) {
	return c.Call.Action()
}

func (c *Call) Cancel() {
	if c.reported.Swap(true) {
		return
	}
	c.Call.Cancel()
}

func (c *Call) Result(ok json.RawMessage, err error) {
	if c.reported.Swap(true) {
		return
	}
	c.Call.Result(ok, err)
}
