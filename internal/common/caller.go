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
	"fmt"
	"runtime"
	"strings"
)

const (
	modulePath = "github.com/doors-dev/doors"
	goxPath    = "github.com/doors-dev/gox"
)

type Caller struct {
	pcs [16]uintptr
	n   int
}

func CaptureCaller() Caller {
	var c Caller
	c.n = runtime.Callers(2, c.pcs[:])
	return c
}

func (c Caller) String() string {
	if c.n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(c.pcs[:c.n])
	for {
		frame, more := frames.Next()
		if !internalFrame(frame.Function) {
			return fmt.Sprintf("%s:%d", frame.File, frame.Line)
		}
		if !more {
			return ""
		}
	}
}

func internalFrame(function string) bool {
	for _, path := range []string{modulePath, goxPath} {
		if strings.HasPrefix(function, path+".") || strings.HasPrefix(function, path+"/") {
			return true
		}
	}
	return false
}
