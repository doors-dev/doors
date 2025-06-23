package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type CSP struct {
	// nil or empty = 'self'
	DefaultSources      []string
	ScriptStrictDynamic bool
	// Add extra sources besides  imported hashes and nonces
	ScriptSources []string
	// Add extra sources besides  imported hashes and nonces
	StyleSources []string
	// Add extra sources besides self
	ConnectSources []string
	// nill or empty = 'none' (framework handeles via js)
	FormActions []string // nil - none
	// If null empty 'none'
	ObjectSources []string
	// If null empty 'none'
	FrameSources []string
	// If null empty 'none'
	FrameAcestors []string
	// If null empty 'none'
	BaseURIAllow []string
	// If null empy directive will not be added
	ImgSources []string
	// If null empy directive will not be added
	FontSources []string
	// If null empy directive will not be added
	MediaSources []string
	// If null empy directive will not be added
	Sandbox []string
	// If null empy directive will not be added
	WorkerSources []string
	// Inlines ScriptLocal* and StyleLocal* with nonce, impoves first time loading, not recommended. Adds nonce to style and script direcrives in header
	InlineLocal bool
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
	nonce   string
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

func (c *CSPCollector) Nonce() string {
	if c == nil {
		return ""
	}
	return c.nonce
}

func (c *CSPCollector) Generate() string {
	return c.csp.generate(c.styles, c.scripts, c.nonce)
}

func (c *CSP) NewCollector() *CSPCollector {
	if c == nil {
		return nil
	}
	var nonce string
	if c.InlineLocal {
		nonce = RandId()
	}
	return &CSPCollector{
		csp:     c,
		nonce:   nonce,
		styles:  newCollectedCSP(),
		scripts: newCollectedCSP(),
	}
}

func (c *CSP) generate(styleCollected *collectedCSP, scriptCollected *collectedCSP, nonce string) string {
	def := c.simple("default-src", nil, c.DefaultSources, []string{"'self'"})
	connect := c.simple("connect-src", []string{"'self'"}, c.ConnectSources, nil)
	script := c.collected("script-src", scriptCollected, c.ScriptSources, nonce)
	if c.ScriptStrictDynamic {
		script = script + " " + "'strict-dynamic'"
	}
	style := c.collected("style-src", styleCollected, c.StyleSources, nonce)
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
		"sanbox":     c.Sandbox,
		"worker-src": c.WorkerSources,
	}
	for directive := range optional {
		value := c.simple(directive, nil, optional[directive], nil)
		parts = append(parts, value)
	}
	return c.join(parts)
}

func (c *CSP) collected(directive string, collected *collectedCSP, user []string, nonce string) string {
	parts := []string{"'self'"}
	for _, hash := range collected.hashes {
		value := fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(hash))
		parts = append(parts, value)
	}
	for _, source := range collected.sources {
		parts = append(parts, source)
	}
	if c.InlineLocal {
		value := fmt.Sprintf("'nonce-%s'", nonce)
		parts = append(parts, value)
	}
	return c.simple(directive, parts, user, nil)
}

func (c *CSP) simple(directive string, mandatory []string, user []string, def []string) string {
	if user != nil && len(user) == 0 {
		return ""
	}
	hasUser := user != nil
	hasDefault := def != nil && len(def) > 0
	hasMandatory := mandatory != nil && len(mandatory) > 0
	if user == nil && !hasMandatory && !hasDefault {
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
