// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import "encoding/json"

type CallData struct {
	Name    string
	Arg     any
	Payload Writable
}

type Call interface {
	Data() *CallData
	Cancel()
	Result(json.RawMessage, error)
}
