package attr

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

func TestInputFocus(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &inputFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.Click(t, page, "#focus")
	test.TestReportId(t, page, 0, "focus")
	test.TestReportId(t, page, 1, "in")
	test.TestReportId(t, page, 2, "in")
	test.Click(t, page, "#blur")
	test.TestReportId(t, page, 0, "blur")
	test.TestReportId(t, page, 1, "out")
	test.TestReportId(t, page, 2, "in")
}

func TestInputInput(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &inputFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.TestType(t, page, "#input", []input.Key{input.KeyA})
	test.TestReportId(t, page, 0, "a")
	test.TestReportId(t, page, 1, "a")
	test.TestType(t, page, "#input", []input.Key{input.KeyB})
	test.TestReportId(t, page, 0, "b")
	test.TestReportId(t, page, 1, "ab")
	test.TestType(t, page, "#input_ex", []input.Key{input.KeyA})
	test.TestReportId(t, page, 0, "a")
	test.TestReportId(t, page, 1, "")
	test.TestType(t, page, "#input_ex", []input.Key{input.KeyB})
	test.TestReportId(t, page, 0, "b")
	test.TestReportId(t, page, 1, "")

}
func TestInputChange(t *testing.T) {

	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &inputFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	_ = proto.EmulationSetTimezoneOverride{TimezoneID: "UTC"}.Call(page)
	text := map[string]string{
		"password": doors.RandId(),
		"textarea": doors.RandId(),
		"search":   doors.RandId(),
		"text":     doors.RandId(),
		"tel":      fmt.Sprint(time.Now().Nanosecond()),
		"url":      strings.ToLower("https://" + doors.RandId() + ".com"),
		"email":    strings.ToLower(doors.RandId() + "@" + doors.RandId() + ".com"),
	}

	for kind := range text {
		str := text[kind]
		test.TestInput(t, page, "#"+kind, str)
		<-time.After(100 * time.Millisecond)
		test.TestReportId(t, page, 0, kind)
		test.TestReportId(t, page, 1, str)
		test.TestReportId(t, page, 2, "")
		test.TestReportId(t, page, 3, "")
		test.TestReportId(t, page, 4, "")
		test.TestReportId(t, page, 5, "false")
	}

	num := time.Now().Nanosecond()
	str := fmt.Sprint(num)

	test.TestInput(t, page, "#number", str)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "number")
	test.TestReportId(t, page, 1, str)
	test.TestReportId(t, page, 2, str)
	test.TestReportId(t, page, 3, "")
	test.TestReportId(t, page, 4, "")
	test.TestReportId(t, page, 5, "false")

	type date struct {
		Value     string
		Number    string
		Date      string
		ParseDate string
	}

	dates := map[string]date{
		"date": {
			Value:  "2025-08-21",
			Number: "1755734400000",
			Date:   "2025-08-21 00:00:00 +0000 UTC",
		},
		"datetime-local": {
			Value:     "2025-08-16T18:11",
			Number:    "1755367860000",
			Date:      "",
			ParseDate: "2025-08-16 18:11:00 +0000 UTC",
		},
		"month": {
			Value:  "2025-11",
			Number: "670",
			Date:   "2025-11-01 00:00:00 +0000 UTC",
		},
		/*	rod issue
			"week": {
				Value:  "2025-W34",
				Number: "1755475200000",
				Date:   "2025-08-18 00:00:00 +0000 UTC",
			}, */
		"time": {
			Value:  "22:20",
			Number: "80400000",
			Date:   "1970-01-01 22:20:00 +0000 UTC",
		},
	}

	for kind := range dates {
		date := dates[kind]
		var now time.Time
		if date.ParseDate != "" {
			now, _ = time.Parse("2006-01-02 15:04:05 -0700 MST", date.ParseDate)
		} else {
			now, _ = time.Parse("2006-01-02 15:04:05 -0700 MST", date.Date)
		}
		test.TestInputTime(t, page, "#"+kind, now)
		<-time.After(200 * time.Millisecond)
		test.TestReportId(t, page, 0, kind)
		test.TestReportId(t, page, 1, date.Value)
		test.TestReportId(t, page, 2, date.Number)
		test.TestReportId(t, page, 3, date.Date)
		test.TestReportId(t, page, 4, "")
		test.TestReportId(t, page, 5, "false")
	}

	str = "#344323"
	test.TestInputColor(t, page, "#color", str)
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "color")
	test.TestReportId(t, page, 1, str)
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	test.TestReportId(t, page, 4, "")
	test.TestReportId(t, page, 5, "false")

	type clickable struct {
		Target  string
		Kind    string
		Value   string
		Checked string
	}

	clickables := []clickable{{
		Target:  "#checkbox",
		Kind:    "checkbox",
		Value:   "on",
		Checked: "true",
	}, {
		Target:  "#checkbox",
		Kind:    "checkbox",
		Value:   "on",
		Checked: "false",
	}, {
		Target:  "#radio-1",
		Kind:    "radio",
		Value:   "option1",
		Checked: "true",
	}, {
		Target:  "#radio-2",
		Kind:    "radio",
		Value:   "option2",
		Checked: "true",
	}}
	for _, c := range clickables {
		test.ClickNow(t, page, c.Target)
		<-time.After(100 * time.Millisecond)
		test.TestReportId(t, page, 0, c.Kind)
		test.TestReportId(t, page, 1, c.Value)
		test.TestReportId(t, page, 2, "")
		test.TestReportId(t, page, 3, "")
		test.TestReportId(t, page, 4, "")
		test.TestReportId(t, page, 5, c.Checked)
	}

	test.TestSelect(t, page, "#select", []string{"Option 1"})
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "select")
	test.TestReportId(t, page, 1, "option1")
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	test.TestReportId(t, page, 4, "option1")
	test.TestReportId(t, page, 5, "false")

	test.TestSelect(t, page, "#multiselect", []string{"Option 1", "Option 2"})
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "multiselect")
	test.TestReportId(t, page, 1, "option1")
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	test.TestReportId(t, page, 4, "option1,option2")
	test.TestReportId(t, page, 5, "false")

	test.TestDeselect(t, page, "#multiselect", []string{"Option 1"})
	<-time.After(100 * time.Millisecond)
	test.TestReportId(t, page, 0, "multiselect")
	test.TestReportId(t, page, 1, "option2")
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	test.TestReportId(t, page, 4, "option2")
	test.TestReportId(t, page, 5, "false")

}
