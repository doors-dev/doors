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

package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// CSP is the Content-Security-Policy Doors sends on page responses.
type CSP struct {
	// DefaultSources are the default-src values. Nil emits 'self'; empty
	// omits the directive.
	DefaultSources []string

	// ScriptSources are extra script-src values. The directive is always
	// emitted with 'self', plus the import map hash and the URLs of
	// doors.ResourceExternal scripts when the page has them.
	ScriptSources []string

	// ScriptStrictDynamic adds 'strict-dynamic' to script-src.
	ScriptStrictDynamic bool

	// StyleSources are extra style-src values. The directive is always emitted
	// with 'self', plus the URLs of doors.ResourceExternal stylesheets when the
	// page has them.
	StyleSources []string

	// ConnectSources are extra connect-src values. The directive is always
	// emitted with 'self', which Doors needs for its own requests.
	ConnectSources []string

	// FormActions are the form-action values. Nil emits 'none'; empty omits
	// the directive.
	FormActions []string

	// ObjectSources are the object-src values. Nil emits 'none'; empty omits
	// the directive.
	ObjectSources []string

	// FrameSources are the frame-src values. Nil emits 'none'; empty omits
	// the directive.
	FrameSources []string

	// FrameAncestors are the frame-ancestors values. Nil emits 'none'; empty
	// omits the directive.
	FrameAncestors []string

	// BaseURIAllow are the base-uri values. Nil emits 'none'; empty omits the
	// directive.
	BaseURIAllow []string

	// ImgSources are the img-src values. Nil or empty omits the directive.
	ImgSources []string

	// FontSources are the font-src values. Nil or empty omits the directive.
	FontSources []string

	// MediaSources are the media-src values. Nil or empty omits the directive.
	MediaSources []string

	// Sandbox are the sandbox flags. Nil or empty omits the directive.
	Sandbox []string

	// WorkerSources are the worker-src values. Nil or empty omits the
	// directive.
	WorkerSources []string

	// ReportTo is the report-to group name. Empty omits the directive.
	ReportTo string
}

type collectedCSP struct {
	sources []string
	hashes  [][]byte
}

func newCollectedCSP() *collectedCSP {
	return &collectedCSP{
		sources: make([]string, 0),
		hashes:  make([][]byte, 0),
	}
}

// CSPCollector collects the sources and hashes added while one page renders.
// A nil collector discards them.
type CSPCollector = *cspCollector

type cspCollector struct {
	csp     *CSP
	styles  *collectedCSP
	scripts *collectedCSP
}

// StyleSource records a stylesheet URL for the final header.
func (c CSPCollector) StyleSource(source string) {
	if c == nil {
		return
	}
	c.styles.sources = append(c.styles.sources, source)
}

// ScriptSource records a script URL for the final header.
func (c CSPCollector) ScriptSource(source string) {
	if c == nil {
		return
	}
	c.scripts.sources = append(c.scripts.sources, source)
}

// StyleHash records an inline stylesheet hash for the final header.
func (c CSPCollector) StyleHash(hash []byte) {
	if c == nil {
		return
	}
	c.styles.hashes = append(c.styles.hashes, hash)
}

// ScriptHash records an inline script hash for the final header.
func (c CSPCollector) ScriptHash(hash []byte) {
	if c == nil {
		return
	}
	c.scripts.hashes = append(c.scripts.hashes, hash)
}

// Generate returns the final Content-Security-Policy header value.
func (c CSPCollector) Generate() string {
	return c.csp.generate(c.styles, c.scripts)
}

// NewCollector returns a collector for one page render. A nil c returns a nil
// collector.
func (c *CSP) NewCollector() CSPCollector {
	if c == nil {
		return nil
	}
	return &cspCollector{
		csp:     c,
		styles:  newCollectedCSP(),
		scripts: newCollectedCSP(),
	}
}

func (c *CSP) generate(styleCollected *collectedCSP, scriptCollected *collectedCSP) string {
	def := c.simple("default-src", nil, c.DefaultSources, []string{"'self'"})
	connect := c.simple("connect-src", []string{"'self'"}, c.ConnectSources, nil)
	script := c.collected("script-src", scriptCollected, c.ScriptSources)
	if c.ScriptStrictDynamic {
		script = script + " " + "'strict-dynamic'"
	}
	style := c.collected("style-src", styleCollected, c.StyleSources)
	allow := map[string][]string{
		"form-action":     c.FormActions,
		"object-src":      c.ObjectSources,
		"frame-src":       c.FrameSources,
		"frame-ancestors": c.FrameAncestors,
		"base-uri":        c.BaseURIAllow,
	}
	parts := []string{def, connect, script, style}
	for directive := range allow {
		value := c.simple(directive, nil, allow[directive], []string{"'none'"})
		parts = append(parts, value)
	}
	optional := map[string][]string{
		"img-src":    c.ImgSources,
		"font-src":   c.FontSources,
		"media-src":  c.MediaSources,
		"sandbox":    c.Sandbox,
		"worker-src": c.WorkerSources,
	}
	for directive := range optional {
		value := c.simple(directive, nil, optional[directive], nil)
		parts = append(parts, value)
	}
	if c.ReportTo != "" {
		parts = append(parts, "report-to "+c.ReportTo)
	}
	return c.join(parts)
}

func (c *CSP) collected(directive string, collected *collectedCSP, user []string) string {
	parts := []string{"'self'"}
	for _, hash := range collected.hashes {
		value := fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(hash))
		parts = append(parts, value)
	}
	parts = append(parts, collected.sources...)
	return c.simple(directive, parts, user, nil)
}

func (c *CSP) simple(directive string, mandatory []string, user []string, def []string) string {
	hasUser := user != nil
	hasDefault := len(def) > 0
	hasMandatory := len(mandatory) > 0
	if user == nil && !hasMandatory && !hasDefault {
		return ""
	}
	if user != nil && len(user) == 0 && !hasMandatory {
		return ""
	}
	parts := []string{directive}
	if hasMandatory {
		parts = append(parts, mandatory...)
	}
	if hasUser {
		parts = append(parts, user...)
	} else if hasDefault {
		parts = append(parts, def...)
	}
	return strings.Join(parts, " ")

}

func (c *CSP) join(parts []string) string {
	var filtered []string
	for _, s := range parts {
		if s == "" {
			continue
		}
		filtered = append(filtered, s)
	}
	return strings.Join(filtered, "; ")
}
