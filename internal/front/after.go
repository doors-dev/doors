// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import (
	"encoding/json"
	"net/http"
)

type After struct {
	Name string
	Arg  any
}

func (a *After) MarshalJSON() ([]byte, error) {
	arr := []any{a.Name, a.Arg}
	return json.Marshal(arr)
}

func (a *After) Set(h http.Header) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	h.Set("D00r-After", string(b))
	return nil
}
