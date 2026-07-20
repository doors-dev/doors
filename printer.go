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

package doors

import (
	"github.com/doors-dev/doors/internal/app"
	"github.com/doors-dev/doors/internal/common"
)

// PrinterMiddleware wraps the [gox.Printer] Doors uses to emit HTML.
//
// Doors invokes the middleware once per drain: an initial page render or a
// door render cycle (the updated door plus descendant doors rendered in that
// cycle). It must return a non-nil printer, which receives that drain's jobs
// as strictly sequential Send calls in document order; the middleware
// function itself may be invoked concurrently across drains.
//
// Jobs are pooled and single-use: forward each job to next exactly once and
// never retain it (or its Attrs) after Send returns. Use job.Context() for
// scope information. [gox.JobHeadOpen] Attrs stay mutable until the job is
// serialized (they may be nil on no-output [gox.KindContainer] heads). Do not
// modify Doors-managed output (d0* elements, attributes, and rewritten
// resource URLs), and do not inject head-relevant elements into a page render
// stream before the document head is printed. Middleware runs on the render
// hot path and must not block.
//
// Example (count elements, stamp each with its index):
//
//	type stamp struct {
//		next  gox.Printer
//		count int
//	}
//
//	func (s *stamp) Send(j gox.Job) error {
//		if open, ok := j.(*gox.JobHeadOpen); ok && open.Tag != "d0-r" && open.Attrs != nil {
//			s.count++
//			open.Attrs.Get("data-render-index").Set(strconv.Itoa(s.count))
//		}
//		return s.next.Send(j)
//	}
type PrinterMiddleware = common.PrinterMiddleware

// WithPrinterMiddleware installs the printer middleware into an app.
//
// The middleware applies to every page render and door render cycle of the
// app. A nil middleware disables wrapping.
func WithPrinterMiddleware(m PrinterMiddleware) With {
	return withFunc(func(o *app.Options) {
		o.PrinterMiddleware = m
	})
}
