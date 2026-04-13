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

// Package doors builds server-rendered web apps with typed routing, reactive
// state, and server-handled browser interactions.
//
// Most apps start with [NewRouter], register one or more page models with
// [UseModel], and render dynamic fragments with [Door], [Source], and [Beam].
// Event attrs such as [AClick], [ASubmit], and [ALink] connect DOM events,
// forms, and navigation to Go handlers while still producing regular HTML.
//
// For a guided introduction, see the documents embedded in [DocsFS].
package doors
