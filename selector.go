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

package doors

import "github.com/doors-dev/doors/internal/front"

// Selector picks the elements an [Indicator] applies to.
//
// Query selectors search the whole document. Target and parent selectors need
// an event element and select nothing without one.
type Selector = front.Selector

// SelectorTarget selects the event element.
func SelectorTarget() Selector {
	return front.SelectTarget()
}

// SelectorQuery selects the first element in the document matching query.
func SelectorQuery(query string) Selector {
	return front.SelectQuery(query)
}

// SelectorQueryAll selects every element in the document matching query.
func SelectorQueryAll(query string) Selector {
	return front.SelectQueryAll(query)
}

// SelectorQueryParent selects the closest ancestor of the event element
// matching query.
func SelectorQueryParent(query string) Selector {
	return front.SelectQueryParent(query)
}
