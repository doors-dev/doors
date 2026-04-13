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

package imports

import (
	"io/fs"
	"log"
	"testing"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	test.RunMain(func() int {
		preactDir, _ := fs.Sub(preactFS, "preact_src")
		reactDir, _ := fs.Sub(reactFS, "react_src")
		moduleDir, _ := fs.Sub(moduleFS, "module_src")
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")
		var err error
		preactPath, err = cookModule(preactDir)
		if err != nil {
			log.Fatal(err.Error())
		}
		reactPath, err = cookModule(reactDir)
		if err != nil {
			clean()
			log.Fatal(err.Error())
		}
		modulePath, err = copyTemp(moduleDir)
		if err != nil {
			clean()
			log.Fatal(err.Error())
		}
		moduleBundlePath, err = copyTemp(moduleBundleDir)
		if err != nil {
			clean()
			log.Fatal(err.Error())
		}
		browser = test.NewBrowser()
		code := m.Run()
		clean()
		browser.MustClose()
		return code
	})
}
