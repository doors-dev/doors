// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

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
