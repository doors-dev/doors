// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package pipe

import "github.com/doors-dev/doors/internal/front/action"

type Payload interface {
	Payload() action.Payload
	Release()
}

func EmptyPayload() Payload {
	return emptyPayload{}
}

type emptyPayload struct{}

func (e emptyPayload) Payload() action.Payload {
	return action.NewText("")
}

func (e emptyPayload) Release() {}
