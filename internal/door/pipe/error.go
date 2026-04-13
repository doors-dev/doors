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

package pipe

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

func NewError(err error) Error {
	if e, ok := err.(Error); ok {
		return e
	}
	id := common.RandId()
	slog.Error("door rendering/printing error", "error", err, "error_id", id)
	return Error{
		id:  id,
		err: err,
	}
}

type Error struct {
	err error
	id  string
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Release() {

}

func (e Error) Payload() action.Payload {
	buf := &bytes.Buffer{}
	if err := e.Main().Render(context.Background(), buf); err != nil {
		panic("error rendering error")
	}
	return action.NewTextBytes(buf.Bytes())
}

func (e Error) Main() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Init("div"); err != nil {
			return err
		}
		{
			if err := cur.AttrSet("role", "alert"); err != nil {
				return err
			}
			if err := cur.AttrSet("aria-live", "polite"); err != nil {
				return err
			}
			if err := cur.AttrSet("data-fw", "error"); err != nil {
				return err
			}
			if err := cur.Submit(); err != nil {
				return err
			}
			if err := cur.Text(`Component Error. `); err != nil {
				return err
			}
			if err := cur.Init("span"); err != nil {
				return err
			}
			{
				if err := cur.AttrSet("data-fw", "error-id"); err != nil {
					return err
				}
				if err := cur.Submit(); err != nil {
					return err
				}
				if err := cur.Text(fmt.Sprintf(`ID: %s`, e.id)); err != nil {
					return err
				}
			}
			if err := cur.Close(); err != nil {
				return err
			}
			if err := cur.Raw(fmt.Sprintf(`<!-- %s -->`, e.err.Error())); err != nil {
				return err
			}
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return nil
	})
}
