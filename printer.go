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
// Doors invokes the middleware once per drain of buffered render jobs. A
// drain is either the initial page render (the whole document, including
// fully static pages) or a door render cycle: the updated door plus all
// descendant doors rendered in that cycle, delivered together as one payload.
// The printer the middleware returns must be non-nil and receives that
// drain's job stream; returning next unchanged observes nothing. The
// middleware function itself may be invoked concurrently (one page and
// several door render cycles can print at the same time, even within one
// instance), but each returned printer receives strictly sequential Send
// calls on a single goroutine, in final document order, and is never reused
// across drains.
//
// A middleware can read jobs to collect information about the produced HTML
// and mutate [gox.JobHeadOpen] before forwarding — its Attrs handle is live
// until the job is serialized downstream, so attributes can be inspected,
// injected, or removed. Attrs may be nil on [gox.KindContainer] heads, which
// emit no HTML; skip them when counting elements.
//
// Jobs are pooled and single-use: forward each job to next exactly once, or
// drop it explicitly with [gox.Release]; never retain a job, its Attrs, or
// its Ctx after Send returns (copy values out, e.g. with Attrs.Clone). Use
// job.Context() for scope information — every job carries the context of the
// render scope that produced it, so nested door jobs carry their own door's
// context, and helpers such as [SessionContext] work on it. Do not modify
// Doors-managed output: "d0-r" elements, id="d0r..." and data-d0* attributes,
// and resource-rewritten src/href values. Do not inject <head>, <body>,
// <script>, or <link> element jobs into a page render stream before the
// document head has been printed — Doors scans that stream to place its own
// head content. A non-nil error from Send fails the render. Middleware runs
// on the render hot path and must not block.
//
// Example (count elements, stamp each with its index):
//
//	type stamp struct {
//		next  gox.Printer
//		count int
//	}
//
//	func (s *stamp) Send(j gox.Job) error {
//		if open, ok := j.(*gox.JobHeadOpen); ok &&
//			open.Kind != gox.KindContainer && open.Attrs != nil {
//			s.count++
//			open.Attrs.Get("data-render-index").Set(strconv.Itoa(s.count))
//		}
//		return s.next.Send(j)
//	}
//
//	app := doors.NewApp(page, doors.WithPrinterMiddleware(
//		func(next gox.Printer) gox.Printer {
//			return &stamp{next: next}
//		},
//	))
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
