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

package attr

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func moveTo(page *rod.Page, id string) (float64, float64) {
	box := page.MustElement("#" + id)
	xy := box.MustShape().Quads[0]
	page.Mouse.MustMoveTo(xy[0], xy[1])
	return xy[0], xy[1]
}

func TestAttrPointer(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &pointerFragment{
			r: test.NewReporter(3),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestReportId(t, page, 0, "")
	x, y := moveTo(page, "down")
	page.Mouse.MustDown(proto.InputMouseButtonLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "DOWN")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "up")
	test.TestReportId(t, page, 0, "DOWN")
	page.Mouse.MustUp(proto.InputMouseButtonLeft)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "UP")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "enter")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "ENTER")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	moveTo(page, "leave")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "ENTER")
	x, y = moveTo(page, "beforeLeave")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "LEAVE")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "move")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "MOVE")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	x, y = moveTo(page, "over")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "OVER")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
	moveTo(page, "out")
	x, y = moveTo(page, "beforeOut")
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "OUT")
	test.TestReportId(t, page, 1, test.Float(x))
	test.TestReportId(t, page, 2, test.Float(y))
}

func parseFloat(t *testing.T, s string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("parseFloat: %q: %v", s, err)
	}
	return f
}

const epsilon = 5.0

type reportedCoords struct {
	offsetX, offsetY             float64
	clientX, clientY             float64
	pageX, pageY                 float64
	screenX, screenY             float64
	pointerW, pointerH           float64
	targetX, targetY             float64
	targetW, targetH             float64
	pageScrollX, pageScrollY     float64
	pageW, pageH                 float64
	screenOffsetX, screenOffsetY float64
	screenW, screenH             float64
}

func readCoords(t *testing.T, page *rod.Page) reportedCoords {
	return reportedCoords{
		offsetX:       parseFloat(t, test.GetReportContent(t, page, 0)),
		offsetY:       parseFloat(t, test.GetReportContent(t, page, 1)),
		clientX:       parseFloat(t, test.GetReportContent(t, page, 2)),
		clientY:       parseFloat(t, test.GetReportContent(t, page, 3)),
		pageX:         parseFloat(t, test.GetReportContent(t, page, 4)),
		pageY:         parseFloat(t, test.GetReportContent(t, page, 5)),
		screenX:       parseFloat(t, test.GetReportContent(t, page, 6)),
		screenY:       parseFloat(t, test.GetReportContent(t, page, 7)),
		pointerW:      parseFloat(t, test.GetReportContent(t, page, 8)),
		pointerH:      parseFloat(t, test.GetReportContent(t, page, 9)),
		targetX:       parseFloat(t, test.GetReportContent(t, page, 10)),
		targetY:       parseFloat(t, test.GetReportContent(t, page, 11)),
		targetW:       parseFloat(t, test.GetReportContent(t, page, 12)),
		targetH:       parseFloat(t, test.GetReportContent(t, page, 13)),
		pageScrollX:   parseFloat(t, test.GetReportContent(t, page, 14)),
		pageScrollY:   parseFloat(t, test.GetReportContent(t, page, 15)),
		pageW:         parseFloat(t, test.GetReportContent(t, page, 16)),
		pageH:         parseFloat(t, test.GetReportContent(t, page, 17)),
		screenOffsetX: parseFloat(t, test.GetReportContent(t, page, 18)),
		screenOffsetY: parseFloat(t, test.GetReportContent(t, page, 19)),
		screenW:       parseFloat(t, test.GetReportContent(t, page, 20)),
		screenH:       parseFloat(t, test.GetReportContent(t, page, 21)),
	}
}

func verifyInvariants(t *testing.T, c reportedCoords) {
	if math.Abs(c.clientX-(c.offsetX+c.targetX)) > epsilon {
		t.Fatalf("clientX %f != offsetX %f + target.X %f (diff=%f)", c.clientX, c.offsetX, c.targetX, c.clientX-(c.offsetX+c.targetX))
	}
	if math.Abs(c.clientY-(c.offsetY+c.targetY)) > epsilon {
		t.Fatalf("clientY %f != offsetY %f + target.Y %f (diff=%f)", c.clientY, c.offsetY, c.targetY, c.clientY-(c.offsetY+c.targetY))
	}
	if math.Abs(c.pageX-(c.clientX+c.pageScrollX)) > epsilon {
		t.Fatalf("pageX %f != clientX %f + page.scrollX %f (diff=%f)", c.pageX, c.clientX, c.pageScrollX, c.pageX-(c.clientX+c.pageScrollX))
	}
	if math.Abs(c.pageY-(c.clientY+c.pageScrollY)) > epsilon {
		t.Fatalf("pageY %f != clientY %f + page.scrollY %f (diff=%f)", c.pageY, c.clientY, c.pageScrollY, c.pageY-(c.clientY+c.pageScrollY))
	}
	if math.Abs(c.screenX-(c.clientX+c.screenOffsetX)) > epsilon {
		t.Fatalf("screenX %f != clientX %f + screen.X %f (diff=%f)", c.screenX, c.clientX, c.screenOffsetX, c.screenX-(c.clientX+c.screenOffsetX))
	}
	if math.Abs(c.screenY-(c.clientY+c.screenOffsetY)) > epsilon {
		t.Fatalf("screenY %f != clientY %f + screen.Y %f (diff=%f)", c.screenY, c.clientY, c.screenOffsetY, c.screenY-(c.clientY+c.screenOffsetY))
	}
}

func verifyNonZeroDims(t *testing.T, c reportedCoords) {
	if c.targetW <= 0 {
		t.Fatal("target.Width should be > 0")
	}
	if c.targetH <= 0 {
		t.Fatal("target.Height should be > 0")
	}
	if c.pageW <= 0 {
		t.Fatal("page.Width should be > 0")
	}
	if c.pageH <= 0 {
		t.Fatal("page.Height should be > 0")
	}
	if c.screenW <= 0 {
		t.Fatal("screen.Width should be > 0")
	}
	if c.screenH <= 0 {
		t.Fatal("screen.Height should be > 0")
	}
}

func TestAttrPointerCoords(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &pointerCoordsFragment{
			r: test.NewReporter(22),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	box := page.MustElement("#coord-target")
	shape := box.MustShape().Quads[0]
	rodX, rodY := shape[0], shape[1]
	rodW := shape[2] - shape[0]
	rodH := shape[5] - shape[1]

	// --- Click at top-left: offset should be ~0 ---
	page.Mouse.MustMoveTo(rodX, rodY)
	page.Mouse.MustClick(proto.InputMouseButtonLeft)
	<-time.After(200 * time.Millisecond)

	c := readCoords(t, page)

	if math.Abs(c.offsetX) > epsilon {
		t.Fatalf("offsetX %f should be ~0 (clicked at element top-left)", c.offsetX)
	}
	if math.Abs(c.offsetY) > epsilon {
		t.Fatalf("offsetY %f should be ~0 (clicked at element top-left)", c.offsetY)
	}
	if math.Abs(c.targetX-rodX) > epsilon {
		t.Fatalf("target.X %f != rod shape.X %f", c.targetX, rodX)
	}
	if math.Abs(c.targetY-rodY) > epsilon {
		t.Fatalf("target.Y %f != rod shape.Y %f", c.targetY, rodY)
	}
	if math.Abs(c.targetW-rodW) > epsilon {
		t.Fatalf("target.Width %f != rod shape.W %f", c.targetW, rodW)
	}
	if math.Abs(c.targetH-rodH) > epsilon {
		t.Fatalf("target.Height %f != rod shape.H %f", c.targetH, rodH)
	}
	if math.Abs(c.clientX-rodX) > epsilon {
		t.Fatalf("clientX %f != mouse X %f (top-left click)", c.clientX, rodX)
	}
	if math.Abs(c.clientY-rodY) > epsilon {
		t.Fatalf("clientY %f != mouse Y %f (top-left click)", c.clientY, rodY)
	}
	verifyInvariants(t, c)

	// --- Click at center (via MustClick on the element itself): offset should be non-zero ---
	box.MustClick()
	<-time.After(200 * time.Millisecond)

	c = readCoords(t, page)

	if c.offsetX <= epsilon {
		t.Fatalf("offsetX %f should be > 0 (center click)", c.offsetX)
	}
	if c.offsetY <= epsilon {
		t.Fatalf("offsetY %f should be > 0 (center click)", c.offsetY)
	}
	if math.Abs(c.offsetX-c.targetW/2) > epsilon {
		t.Fatalf("offsetX %f should be ~ target.Width/2 %f (center click)", c.offsetX, c.targetW/2)
	}
	if math.Abs(c.offsetY-c.targetH/2) > epsilon {
		t.Fatalf("offsetY %f should be ~ target.Height/2 %f (center click)", c.offsetY, c.targetH/2)
	}
	verifyInvariants(t, c)

	// --- Shared sanity checks ---
	verifyNonZeroDims(t, c)
	if c.pointerW < 0 {
		t.Fatalf("pointer.Width %f should be >= 0", c.pointerW)
	}
	if c.pointerH < 0 {
		t.Fatalf("pointer.Height %f should be >= 0", c.pointerH)
	}
	if c.pageScrollX < 0 {
		t.Fatalf("page.scrollX %f should be >= 0", c.pageScrollX)
	}
	if c.pageScrollY < 0 {
		t.Fatalf("page.scrollY %f should be >= 0", c.pageScrollY)
	}
}
