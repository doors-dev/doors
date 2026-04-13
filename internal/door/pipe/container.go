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
	"context"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type ProxyContainer struct {
	Tag   string
	Attrs gox.Attrs
}

func (p ProxyContainer) Apply(pipe Pipe, containerCtx context.Context, doorID uint64, parentID uint64) {
	headID := pipe.cursor.NewID()
	var openJob *gox.JobHeadOpen
	var closeJob *gox.JobHeadClose
	if p.Tag == "" {
		attrs := gox.NewAttrs()
		front.AttrsSetDoor(attrs, doorID, true)
		front.AttrsSetParent(attrs, parentID)
		openJob = gox.NewJobHeadOpen(containerCtx, headID, gox.KindRegular, "d0-r", attrs)
		closeJob = gox.NewJobHeadClose(containerCtx, headID, gox.KindRegular, "d0-r")
	} else {
		attrs := p.Attrs.Clone()
		front.AttrsSetDoor(attrs, doorID, false)
		front.AttrsSetParent(attrs, parentID)
		openJob = gox.NewJobHeadOpen(containerCtx, headID, gox.KindRegular, p.Tag, attrs)
		closeJob = gox.NewJobHeadClose(containerCtx, headID, gox.KindRegular, p.Tag)
	}
	pipe.unshift(openJob)
	pipe.push(closeJob)
}
