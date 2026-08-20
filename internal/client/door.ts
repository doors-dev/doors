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

import { rootId } from './params'
import { doorId } from './lib'

import { attach as attachCaptures } from "./capture"
import { attach as attachEmitter } from "./emitter"
import { HookErr } from "./hook_err"
import navigator from "./navigator"
import { attach as attachSetter } from "./setter"

type Handler = ((arg: any) => any) | ((arg: any, err: HookErr) => any)
type Closure = () => void | Promise<void>

const attr = "data-d0r"
const tag = "d0-r"

const doorState = Symbol()

type DoorElement = Element & {
	[doorState]: {
		id: number
		parent: number
	}
}

function execute(c: Closure) {

	const logError = (e: any) => {
		console.error("unmount handler error", e)
	}
	try {
		const result = c()
		if (result instanceof Promise) {
			result.then().catch(e => logError(e))
		}
	} catch (e) {
		logError(e)
	}
}
function getSelfId(el: Element): (number | undefined) {
	if (el.matches(`${tag}`)) {
		return doorId(el.id)
	}
	if (el.matches(`[${attr}]`)) {
		const id = el.getAttribute(attr)
		return Number(id)
	}
	return undefined
}

function getParentId(el: Element): number {
	const parentAttr = el.getAttribute("data-d0p")
	if (parentAttr !== null) {
		return Number(parentAttr)
	}
	const parent = el.parentElement?.closest(`${tag}, [${attr}]`)
	if (!parent) {
		return rootId
	}
	return getSelfId(parent)!
}



class Doors {
	private elements = new Map<number, DoorElement>()
	private handlers = new Map<number, Map<string, Handler>>()
	private onClear = new Map<number, Array<Closure>>()
	private onRemove = new Map<number, Array<Closure>>()
	private children = new Map<number, Set<number>>()

	private scanDoors(parent: Element | Document | DocumentFragment) {
		for (const element of parent.querySelectorAll<Element>(`${tag}, [${attr}]`)) {
			if (doorState in element) {
				continue
			}
			this.register(element)
		}
	}

	private clear(id: number) {
		this.handlers.delete(id)
		this.clearClosures(id)
		this.clearChildren(id)
	}

	private clearClosures(id: number) {
		const closures = this.onClear.get(id)
		if (!closures) {
			return
		}
		this.onClear.delete(id)
		closures.forEach(c => execute(c))
	}

	private clearChildren(id: number) {
		const children = this.children.get(id)
		if (!children) {
			return
		}
		this.children.delete(id)
		for (const child of children) {
			const element = this.elements.get(child)!
			this.unregister(element)
		}
	}

	register(element: Element) {
		const door = element as DoorElement
		door[doorState] = {
			id: getSelfId(element)!,
			parent: getParentId(element),
		}
		this.elements.set(door[doorState].id, door)
		let siblings = this.children.get(door[doorState].parent)
		if (!siblings) {
			siblings = new Set()
			this.children.set(door[doorState].parent, siblings)
		}
		siblings.add(door[doorState].id)
	}

	unregister(element: Element): void {
		const door = element as DoorElement
		this.elements.delete(door[doorState].id)
		this.clear(door[doorState].id)
		const onRemove = this.onRemove.get(door[doorState].id)
		if (onRemove !== undefined) {
			this.onRemove.delete(door[doorState].id)
			onRemove.forEach(c => execute(c))
		}
		const siblings = this.children.get(door[doorState].parent)
		if (!siblings) {
			return
		}
		siblings.delete(door[doorState].id)
		if (siblings.size == 0) {
			this.children.delete(door[doorState].parent)
		}
	}

	scan(parent: Element | Document | DocumentFragment) {
		this.scanDoors(parent)
		attachEmitter(parent)
		attachCaptures(parent)
		attachSetter(parent)
		navigator.scan(parent)
	}

	update(id: number, content: string) {
		const door = this.elements.get(id)
		if (!door) {
			throw new Error(`door ${id} not found`)
		}
		this.clear(door[doorState].id)

		const range = document.createRange()
		range.selectNodeContents(door)
		range.deleteContents()
		const fragment = range.createContextualFragment(content)
		this.scan(fragment)
		range.insertNode(fragment)
	}

	replace(id: number, content: string) {
		const door = this.elements.get(id)
		if (!door) {
			throw new Error(`door ${id} not found`)
		}
		this.unregister(door)
		const range = document.createRange()
		range.selectNode(door)
		range.deleteContents()
		const fragment = range.createContextualFragment(content)
		this.scan(fragment)
		range.insertNode(fragment)
	}

	freeze(id: number) {
		const door = this.elements.get(id)
		if (!door) {
			throw new Error(`door ${id} not found`)
		}
		this.unregister(door)
	}

	on(
		element: Element,
		name: string,
		handler: Handler,
	): void {
		let id = getSelfId(element)
		if (id == undefined) {
			id = getParentId(element)
		}
		let handlers = this.handlers.get(id)
		if (!handlers) {
			handlers = new Map()
			this.handlers.set(id, handlers)
		}
		handlers.set(name, handler)
	}

	onUnmount(element: Element, handler: Closure): void {
		let id = getSelfId(element)
		let index = this.onRemove
		if (id === undefined) {
			id = getParentId(element)
			index = this.onClear
		}
		const closures = index.get(id)
		if (!closures) {
			index.set(id, [handler])
			return
		}
		closures.push(handler)
	}

	getHandler(id: number, name: string): Handler | undefined {
		const handlers = this.handlers.get(id)
		if (handlers !== undefined && handlers.has(name)) {
			return handlers.get(name)
		}
		if (id == rootId) {
			return undefined
		}
		const element = this.elements.get(id)
		if (element === undefined) {
			console.error(`unexpected behavior: door [${id}] not found for call [${name}]`)
			return undefined
		}
		return this.getHandler(element[doorState].parent, name)
	}

}

const doors = new Doors()

export default doors
