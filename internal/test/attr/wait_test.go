package attr

import (
	"fmt"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func waitContent(t *testing.T, page *rod.Page, selector string, content string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if text, ok := textContent(page, selector); ok && text == content {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestContent(t, page, selector, content)
}

func waitReportId(t *testing.T, page *rod.Page, id int, content string, timeout time.Duration) {
	t.Helper()
	waitContent(t, page, fmt.Sprintf("#report-%d", id), content, timeout)
}

func waitAttr(t *testing.T, page *rod.Page, selector string, name string, value string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		attr, ok := attrValue(page, selector, name)
		if ok && attr == value {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestAttr(t, page, selector, name, value)
}

func waitAttrNo(t *testing.T, page *rod.Page, selector string, name string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, ok := attrValue(page, selector, name); !ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestAttrNo(t, page, selector, name)
}

func waitAttrNot(t *testing.T, page *rod.Page, selector string, name string, value string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		attr, ok := attrValue(page, selector, name)
		if !ok || attr != value {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestAttrNot(t, page, selector, name, value)
}

func attrValue(page *rod.Page, selector string, name string) (string, bool) {
	el, err := page.Timeout(50 * time.Millisecond).Element(selector)
	if err != nil {
		return "", false
	}
	attr, err := el.Attribute(name)
	if err != nil || attr == nil {
		return "", false
	}
	return *attr, true
}

func textContent(page *rod.Page, selector string) (string, bool) {
	el, err := page.Timeout(50 * time.Millisecond).Element(selector)
	if err != nil {
		return "", false
	}
	text, err := el.Text()
	if err != nil {
		return "", false
	}
	return text, true
}
