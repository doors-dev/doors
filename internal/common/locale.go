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
	"time"

	"golang.org/x/text/language"
)

type Locale struct {
	// Language is the browser's language. It is [language.Und] before the
	// session's first page is served, for example in middleware, or when neither
	// the request nor the browser provides a usable language.
	Language language.Tag
	// TimeZone is the browser's time zone. It is [time.Local] until the browser
	// reports a zone the server knows.
	TimeZone *time.Location
	// Synced reports whether the browser has reported its language and time
	// zone in this session. It turns true with the first report and stays true.
	Synced bool
}
