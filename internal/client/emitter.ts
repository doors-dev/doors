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

import doors from "./door"

class Registry {
	private registry_ = new Map<number, Set<Element>>()

	add(element: Element, id: number) {
		let elements = this.registry_.get(id)
		if (!elements) {
			elements = new Set<Element>()
			this.registry_.set(id, elements)
		}
		if (elements.has(element)) {
			return
		}
		elements.add(element)
		doors.onUnmount(element, () => {
			elements.delete(element)
			if (elements.size === 0) {
				this.registry_.delete(id)
			}
		})
	}
	get(id: number): Set<Element> | undefined {
		return this.registry_.get(id)
	}
}


const r = new Registry()


export function getEmitter(id: number): Set<Element> | undefined {
	return r.get(id)
}

const attr = "data-d0e"
export function attach(parent: Element | DocumentFragment | Document) {
	for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="applied"])`)) {
		const ids = JSON.parse(element.getAttribute(attr)!)
		element.setAttribute(attr, "applied")
		for (const id of ids) {
			r.add(element, id)
		}
	}
}
