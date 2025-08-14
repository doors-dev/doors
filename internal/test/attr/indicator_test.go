package attr

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
)

func TestIndicatorSelectors(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &indicatorFragment{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()


	test.TestAttrNot(t, page, "#parent", "data-check", "true")
	test.ClickNow(t, page, "#indicate-parent")
	<-time.After(20 * time.Millisecond)
	test.TestAttr(t, page, "#parent", "data-check", "true")
	<-time.After(500 * time.Millisecond)
	test.TestAttrNot(t, page, "#parent", "data-check", "true")

	test.ClickNow(t, page, "#indicate-self")
	<-time.After(20 * time.Millisecond)
	test.TestContent(t, page, "#indicate-self", "indication")
	<-time.After(500 * time.Millisecond)
	test.TestContent(t, page, "#indicate-self", "indicate-self")

	test.TestAttrNot(t, page, "#next", "data-check", "true")
	test.ClickNow(t, page, "#indicate-selector")
	<-time.After(20 * time.Millisecond)
	test.TestAttr(t, page, "#next", "data-check", "true")
	<-time.After(500 * time.Millisecond)
	test.TestAttrNot(t, page, "#next", "data-check", "true")
}

func TestIndicatorRestore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &indicatorFragment{}
	})
	page := bro.Page(t, "/")

	// initial state
	test.TestAttr(t, page, "#indicator-1", "data-attr1", "val-1")
	test.TestAttrNo(t, page, "#indicator-1", "data-attr2")
	test.TestClass(t, page, "#indicator-1", "class-1")
	test.TestClass(t, page, "#indicator-1", "class-3")
	test.TestClassNot(t, page, "#indicator-1", "class-2")
	test.TestContent(t, page, "#indicator-1", "content-1")

	// during indication
	test.ClickNow(t, page, "#action-1")
	<-time.After(20 * time.Millisecond)
	test.TestAttr(t, page, "#indicator-1", "data-attr1", "val-other")
	test.TestClassNot(t, page, "#indicator-1", "class-3")
	test.TestAttr(t, page, "#indicator-1", "data-attr2", "val-2")
	test.TestClass(t, page, "#indicator-1", "class-1")
	test.TestClass(t, page, "#indicator-1", "class-2")
	test.TestContent(t, page, "#indicator-1", "indication")

	// after indication cleared
	<-time.After(500 * time.Millisecond)
	test.TestAttr(t, page, "#indicator-1", "data-attr1", "val-1")
	test.TestClassNot(t, page, "#indicator-1", "class-1")
	test.TestClass(t, page, "#indicator-1", "class-3")
	test.TestClassNot(t, page, "#indicator-1", "class-2")
	test.TestContent(t, page, "#indicator-1", "content-1")
	test.TestAttrNo(t, page, "#indicator-1", "data-attr2")
}

func TestIndicatorQueue(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &indicatorFragment{}
	})
	page := bro.Page(t, "/")

	// Initial
	test.TestContent(t, page, "#q-target", "base")
	test.TestClass(t, page, "#q-target", "base-class")
	test.TestClassNot(t, page, "#q-target", "class-1")
	test.TestClassNot(t, page, "#q-target", "class-2")
	test.TestAttr(t, page, "#q-target", "data-a", "A0")
	test.TestAttrNo(t, page, "#q-target", "data-b")

	// Start first indication
	test.ClickNow(t, page, "#queue-1")
	test.TestContent(t, page, "#q-target", "first")
	test.TestClass(t, page, "#q-target", "class-1")
	test.TestClassNot(t, page, "#q-target", "class-2")
	test.TestAttr(t, page, "#q-target", "data-a", "A1")
	test.TestAttrNo(t, page, "#q-target", "data-b")

	<-time.After(300 * time.Millisecond)
	test.ClickNow(t, page, "#queue-2")
	// Still first active
	test.TestContent(t, page, "#q-target", "first")
	test.TestClass(t, page, "#q-target", "class-1")
	test.TestClassNot(t, page, "#q-target", "class-2")
	test.TestAttr(t, page, "#q-target", "data-a", "A1")
	test.TestAttrNo(t, page, "#q-target", "data-b")

	// After first completes, second applies (partial update behavior)
	<-time.After(250 * time.Millisecond)
	test.TestContent(t, page, "#q-target", "second")
	test.TestClassNot(t, page, "#q-target", "class-1")
	test.TestClass(t, page, "#q-target", "class-2")
	test.TestAttr(t, page, "#q-target", "data-a", "A0")
	test.TestAttr(t, page, "#q-target", "data-b", "B2")
	// Queue second while first is active

	test.ClickNow(t, page, "#queue-3")
	test.TestClass(t, page, "#q-target", "class-2")
	test.TestContent(t, page, "#q-target", "second")
	test.TestAttr(t, page, "#q-target", "data-a", "A0")
	test.TestAttr(t, page, "#q-target", "data-b", "B2")
	<-time.After(300 * time.Millisecond)
	test.TestClass(t, page, "#q-target", "class-2")
	test.TestClass(t, page, "#q-target", "class-3")
	test.TestAttr(t, page, "#q-target", "data-b", "B2")
	test.TestContent(t, page, "#q-target", "second")
	test.TestAttr(t, page, "#q-target", "data-a", "A3")

	// After thirs completes, restore original
	<-time.After(200 * time.Millisecond)
	test.TestContent(t, page, "#q-target", "base")
	test.TestClass(t, page, "#q-target", "base-class")
	test.TestClassNot(t, page, "#q-target", "class-1")
	test.TestClassNot(t, page, "#q-target", "class-3")
	test.TestClassNot(t, page, "#q-target", "class-2")
	test.TestAttr(t, page, "#q-target", "data-a", "A0")
	test.TestAttrNo(t, page, "#q-target", "data-b")
}
