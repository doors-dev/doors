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

package printer

import (
	"fmt"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type resourceTag interface {
	Read(gox.Attrs) (bool, error)
	Validate() error
	Submit(*gox.JobOpen, *resourcePrinter) error
}

func newAssetTag(tag string) resourceTag {
	r := &assetTag{}
	r.tagBase = tagBase{target: &r.Asset, tag: tag}
	return r
}

type tagBase struct {
	target       Resource
	tag          string
	attrsToClean []gox.Attr
	sourceAttr   gox.Attr
}

type assetTag struct {
	Asset
	tagBase
}

func (r *assetTag) Read(attrs gox.Attrs) (bool, error) {
	attr, ok := attrs.Find("src")
	if !ok {
		attr, ok = attrs.Find("href")
		if !ok {
			return false, nil
		}
	}
	if err := r.setSource(attr); err != nil {
		return true, err
	}
	if r.srcKind == sourceUnset {
		panic("internal error: source kind is unset after reading a resource source")
	}
	if r.srcKind == sourceUnknown {
		return false, nil
	}
	if r.srcKind == sourceLink {
		return false, nil
	}
	for attr := range attrs.All() {
		match, err := r.readMode(attr)
		if err != nil {
			return true, err
		}
		if match {
			continue
		}
		if r.readName(attr) {
			continue
		}
		if attr.Name() == "type" && attr.IsSet() {
			contentType, _ := attr.Value().(string)
			r.SetContentType(contentType)
		}
	}
	return true, nil
}

func (r *assetTag) Submit(openJob *gox.JobOpen, p *resourcePrinter) error {
	r.cleanAttrs()
	c := openJob.Context().Value(common.KeyCore).(core.Core)
	url, err := r.Resolve(c)
	if err != nil {
		return err
	}
	r.sourceAttr.Set(url)
	return p.printer.Send(openJob)
}

func (t *tagBase) reg(a gox.Attr) {
	t.attrsToClean = append(t.attrsToClean, a)
}

func (t *tagBase) cleanAttrs() {
	for _, attr := range t.attrsToClean {
		attr.Unset()
	}
	t.attrsToClean = nil
}

func (t *tagBase) setSource(attr gox.Attr) error {
	t.sourceAttr = attr
	if err := t.target.SetSource(attr.Value()); err != nil {
		return fmt.Errorf("%w on <%s>", err, t.tag)
	}
	return nil
}

func (t *tagBase) readString(attr gox.Attr, name string, set func(string)) bool {
	if attr.Name() != name {
		return false
	}
	value, ok := attr.Value().(string)
	if !ok {
		return true
	}
	set(value)
	t.reg(attr)
	return true
}

func (t *tagBase) readName(attr gox.Attr) bool {
	return t.readString(attr, "name", t.target.SetName)
}

func (t *tagBase) readMode(attr gox.Attr) (bool, error) {
	if attr.Name() == "portable" && isTrue(attr) {
		t.target.SetPortable(true)
		t.reg(attr)
		return true, nil
	}
	var mode resources.ResourceMode
	switch true {
	case attr.Name() == "private" && isTrue(attr):
		mode = resources.ModeNoHost
	case attr.Name() == "nocache" && isTrue(attr):
		mode = resources.ModeNoCache
	case attr.Name() == "cache" && isTrue(attr):
		mode = resources.ModeHost
	default:
		return false, nil
	}
	if err := t.target.SetMode(mode); err != nil {
		return true, err
	}
	t.reg(attr)
	return true, nil
}
