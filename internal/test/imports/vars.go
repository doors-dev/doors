package imports

import (
	"embed"
	_ "embed"
)

//go:embed preact_src
var preactFS embed.FS

//go:embed react_src
var reactFS embed.FS

//go:embed module_src
var moduleFS embed.FS

//go:embed module_bundle_src
var moduleBundleFS embed.FS

//go:embed module_src
var moduleRawFS embed.FS

var preactPath string
var reactPath string
var modulePath string
var bundleModulePath string
var moduleBundlePath string

//go:embed module_src/index.ts
var moduleBytes []byte

//go:embed module_src/index.js
var moduleRawBytes []byte

//go:embed module_src/style.css
var styleRawBytes []byte
