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

import "github.com/doors-dev/doors"
import "github.com/doors-dev/doors/internal/test"
import "github.com/doors-dev/doors/internal/common"
import "github.com/doors-dev/gox"
import "testing"
import "time"
import "fmt"

func TestAttrData(t *testing.T) {
	data := common.RandId()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &dataFragment{
			data: data,
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	test.TestContent(t, page, "#target", data)
}

func TestAttrHook(t *testing.T) {
	data := common.RandId()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &hookFragment{
			data: data,
			r:    test.NewReporter(2),
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	test.TestContent(t, page, "#target", fmt.Sprint(len(data)))
	test.TestContent(t, page, "#target2", fmt.Sprint(len(data)))
	test.TestReport(t, page, data)
	test.TestReportId(t, page, 1, data)
}
func TestAttrRequestTimeout(t *testing.T) {
	conf := doors.Conf{}
	conf.RequestTimeout = time.Second
	bro := test.NewPathBroOptions(browser, func(s test.PathLens) gox.Comp {
		return &test.Page{
			Source: s,
			F: &timeoutFragment{
				r: test.NewReporter(1),
			},
		}
	}, []doors.With{doors.WithConf(conf)})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	waitContent(t, page, "#slow-default", "err", 2*time.Second)
	waitContent(t, page, "#slow-long", "ok:done", 3*time.Second)
	page.MustElement("#slow-form-submit").MustClick()
	waitReportId(t, page, 0, "submitted", 4*time.Second)
}

func TestAttrCall(t *testing.T) {
	data := common.RandId()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &callFragment{
			data: data,
			r:    test.NewReporter(3),
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	waitContent(t, page, "#target", fmt.Sprint(len(data)), time.Second)
	waitReportId(t, page, 0, data, time.Second)
	waitReportId(t, page, 1, "response", time.Second)
	waitReportId(t, page, 2, "async:"+data, 2*time.Second)
}
