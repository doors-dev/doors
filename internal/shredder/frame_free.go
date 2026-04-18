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

package shredder

import "context"

type FreeFrame struct{}

func (f FreeFrame) Release() {

}

func (f FreeFrame) Run(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(run{runtime: r, ctx: ctx, fun: fun})
}

func (f FreeFrame) Submit(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: r, ctx: ctx, fun: fun})
}

func (f FreeFrame) schedule(e executable) {
	e.execute(func(error) {})
}

var _ Frame = FreeFrame{}
