// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package resources

import (
	"bytes"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
)

type HTMLElement struct {
	Tag     string
	Content []byte
	Attrs   templ.Attributes
}

func HTMLParseElement(tag string, data []byte) (*HTMLElement, error) {
	reader := bytes.NewReader(data)
	door, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}
	content, attrs := destructFirst(door, tag)
	if content == nil {
		return nil, nil
	}
	return &HTMLElement{
		Tag:     tag,
		Content: content,
		Attrs:   attrs,
	}, nil
}

func destructFirst(n *html.Node, tag string) ([]byte, templ.Attributes) {
	if n.Type == html.ElementNode && n.Data == tag {
		var b = &bytes.Buffer{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				b.WriteString(c.Data)
			}
		}
		attrs := make(templ.Attributes)
		for _, attr := range n.Attr {
			attrs[attr.Key] = attr.Val
		}
		return b.Bytes(), attrs
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if content, attrs := destructFirst(c, tag); content != nil {
			return content, attrs
		}
	}
	return nil, nil
}
