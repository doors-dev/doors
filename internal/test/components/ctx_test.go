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

package components

import (
	"testing"

	"github.com/doors-dev/doors/internal/test"
)

func TestCtxValueInSubtree(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#inside", "data-value", "v1")
	test.TestAttr(t, page, "#outside", "data-value", "none")
}

func TestCtxOverrideDerivedFromScope(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#override", "data-value", "v2")
	test.TestAttr(t, page, "#inside", "data-value", "v1")
}

func TestCtxValueCrossesDoorBoundary(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#initial", "data-value", "v1")
	test.TestAttr(t, page, "#initial-deep", "data-value", "v1")
}

func TestCtxValueSurvivesInnerUpdate(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#update-inner")
	test.TestAttr(t, page, "#updated", "data-value", "v1")
	test.TestAttr(t, page, "#updated-deep", "data-value", "v1")
}

func TestCtxValueSurvivesStatic(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#stat-initial", "data-value", "v1")
	test.Click(t, page, "#static-stat")
	test.TestAttr(t, page, "#stat-static", "data-value", "v1")
}

func TestCtxCanceledScopeStillRenders(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#canceled", "data-value", "vc")
}

func TestCtxRerenderUsesNewValue(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestAttr(t, page, "#dyn-a", "data-value", "a")
	test.Click(t, page, "#rerender-wrap")
	test.TestAttr(t, page, "#dyn-b", "data-value", "b")
	test.TestMustNot(t, page, "#dyn-a")
}

func TestCtxContainerHandlerSeesValue(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#dyn-el")
	test.TestAttr(t, page, "#container-report", "data-value", "a")
}

func TestCtxContentHandlerSeesValue(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#content-handler")
	test.TestAttr(t, page, "#content-report", "data-value", "a")
}

func TestCtxContainerHandlerSeesNewValueAfterRerender(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#rerender-wrap")
	test.Click(t, page, "#dyn-el")
	test.TestAttr(t, page, "#container-report", "data-value", "b")
}

func TestCtxContainerPreservedAcrossInner(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#update-dyn")
	test.TestAttr(t, page, "#dyn-updated", "data-value", "a")
	test.Click(t, page, "#dyn-el")
	test.TestAttr(t, page, "#container-report", "data-value", "a")
}

func TestCtxSameScopeHandlerGetsTrackerValues(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &CtxRerenderFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.Click(t, page, "#same-scope-handler")
	test.TestAttr(t, page, "#same-scope-report", "data-value", "none")
}
