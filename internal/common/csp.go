// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type CSP struct {
	// default-src:
	//   nil          ⇒ emit:   default-src 'self'
	//   len == 0     ⇒ omit:   (no default-src directive)
	//   len > 0      ⇒ emit:   default-src <values>
	DefaultSources []string

	// script-src (user additions only):
	//   Directive is ALWAYS emitted with at least 'self' plus any collected hashes/sources.
	//   nil or len == 0 ⇒ no extra user sources added.
	ScriptSources []string

	// script-src strict-dynamic:
	//   When true, appends 'strict-dynamic' to script-src (typically used with nonces/hashes).
	ScriptStrictDynamic bool

	// style-src (user additions only):
	//   Directive is ALWAYS emitted with at least 'self' plus any collected hashes/sources.
	//   nil or len == 0 ⇒ no extra user sources added.
	StyleSources []string

	// connect-src (user additions only):
	//   Directive is ALWAYS emitted with at least 'self'.
	//   nil or len == 0 ⇒ results in: connect-src 'self'
	//   len > 0         ⇒ connect-src 'self' <values>
	ConnectSources []string

	// form-action:
	//   nil      ⇒ emit:   form-action 'none'
	//   len == 0 ⇒ omit:   (no form-action directive)
	//   len > 0  ⇒ emit:   form-action <values>
	FormActions []string

	// object-src:
	//   nil      ⇒ emit:   object-src 'none'
	//   len == 0 ⇒ omit:   (no object-src directive)
	//   len > 0  ⇒ emit:   object-src <values>
	ObjectSources []string

	// frame-src:
	//   nil      ⇒ emit:   frame-src 'none'
	//   len == 0 ⇒ omit:   (no frame-src directive)
	//   len > 0  ⇒ emit:   frame-src <values>
	FrameSources []string

	// frame-ancestors:
	//   nil      ⇒ emit:   frame-ancestors 'none'
	//   len == 0 ⇒ omit:   (no frame-ancestors directive)
	//   len > 0  ⇒ emit:   frame-ancestors <values>
	FrameAcestors []string

	// base-uri:
	//   nil      ⇒ emit:   base-uri 'none'
	//   len == 0 ⇒ omit:   (no base-uri directive)
	//   len > 0  ⇒ emit:   base-uri <values>
	BaseURIAllow []string

	// img-src:
	//   nil or len == 0 ⇒ omit
	//   len > 0         ⇒ emit: img-src <values>
	ImgSources []string

	// font-src:
	//   nil or len == 0 ⇒ omit
	//   len > 0         ⇒ emit: font-src <values>
	FontSources []string

	// media-src:
	//   nil or len == 0 ⇒ omit
	//   len > 0         ⇒ emit: media-src <values>
	MediaSources []string

	// sandbox:
	//   nil or len == 0 ⇒ omit
	//   len > 0         ⇒ emit: sandbox <flags>
	Sandbox []string

	// worker-src:
	//   nil or len == 0 ⇒ omit
	//   len > 0         ⇒ emit: worker-src <values>
	WorkerSources []string

	// report-to:
	//   ""  ⇒ omit
	//   set ⇒ emit: report-to <value>
	// NOTE: To make this effective, you must also send a corresponding
	//       `Report-To` HTTP response header that defines the reporting group.
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

type CSPCollector struct {
	csp     *CSP
	styles  *collectedCSP
	scripts *collectedCSP
}

func (c *CSPCollector) StyleSource(source string) {
	if c == nil {
		return
	}
	c.styles.sources = append(c.styles.sources, source)
}

func (c *CSPCollector) ScriptSource(source string) {
	if c == nil {
		return
	}
	c.scripts.sources = append(c.scripts.sources, source)
}

func (c *CSPCollector) StyleHash(hash []byte) {
	if c == nil {
		return
	}
	c.styles.hashes = append(c.styles.hashes, hash)
}

func (c *CSPCollector) ScriptHash(hash []byte) {
	if c == nil {
		return
	}
	c.scripts.hashes = append(c.scripts.hashes, hash)
}

func (c *CSPCollector) Generate() string {
	return c.csp.generate(c.styles, c.scripts)
}

func (c *CSP) NewCollector() *CSPCollector {
	if c == nil {
		return nil
	}
	return &CSPCollector{
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
		"frame-ancestors": c.FrameAcestors,
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
