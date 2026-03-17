// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import { rootId } from './params'
import { doorId } from './lib'

import { attach as attachCaptures, HookErr } from "./capture"
import navigator from "./navigator"
import { attach as attachDyna } from "./dyna"

type Handler = ((arg: any) => any) | ((arg: any, err: HookErr) => any)
type Closure = () => void | Promise<void>


const attr = "data-d0r"
const attrIndexed = "data-d0"
const tag = "d0-r"

type DoorElement = Element & {
	_d0r: {
		id: number
		parent: number
		impostor: boolean
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
	private impostors = new Map<number, Set<number>>()

	private scanImpostors(parent: Element | Document | DocumentFragment) {
		for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attrIndexed}="indexed"])`)) {
			const id = element.getAttribute(attr)
			element.setAttribute(attrIndexed, "indexed")
			this.register(element, { impostorId: id! })
		}
	}

	private clear(id: number) {
		this.handlers.delete(id)
		this.clearClosures(id)
		this.clearImpostors(id)
	}

	private clearClosures(id: number) {
		const closures = this.onClear.get(id)
		if (!closures) {
			return
		}
		this.onClear.delete(id)
		closures.forEach(c => execute(c))
	}

	private clearImpostors(id: number) {
		const impostors = this.impostors.get(id)
		if (!impostors) {
			return
		}
		for (const impostor of impostors) {
			const element = this.elements.get(impostor)!
			this.unregister(element)
		}
	}

	register(element: Element, info: { impostorId?: string }) {
		const impostor = info.impostorId !== undefined;

		const door = element as DoorElement
		const id = impostor ? Number(info.impostorId) : doorId(element.id);
		door._d0r = {
			id: id,
			parent: getParentId(element),
			impostor: impostor,
		}
		this.elements.set(id, door)
		if (!impostor) {
			return
		}
		let siblings = this.impostors.get(door._d0r.parent)
		if (!siblings) {
			siblings = new Set()
			this.impostors.set(door._d0r.parent, siblings)
		}
		siblings.add(id)
	}

	unregister(element: Element): void {
		const door = element as DoorElement
		this.elements.delete(door._d0r.id)
		this.clear(door._d0r.id)
		const onRemove = this.onRemove.get(door._d0r.id)
		if (onRemove !== undefined) {
			this.onRemove.delete(door._d0r.id)
			onRemove.forEach(c => execute(c))
		}
		if (!door._d0r.impostor) {
			return
		}
		const siblings = this.impostors.get(door._d0r.parent)!
		siblings.delete(door._d0r.id)
		if (siblings.size == 0) {
			this.impostors.delete(door._d0r.parent)
		}
	}

	scan(parent: Element | Document | DocumentFragment) {
		this.scanImpostors(parent)
		attachCaptures(parent)
		attachDyna(parent)
		navigator.scan(parent)
	}

	update(id: number, content: string) {
		const door = this.elements.get(id)
		if (!door) {
			throw new Error(`door ${id} not found`)
		}
		this.clear(door._d0r.id)

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
		if (door._d0r.impostor) {
			this.unregister(door)
		}
		const range = document.createRange()
		range.selectNode(door)
		range.deleteContents()
		const fragment = range.createContextualFragment(content)
		this.scan(fragment)
		range.insertNode(fragment)
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



	onUnmount(element: Element, handler: () => void | Promise<void>): void {
		let id = getSelfId(element)
		if (id !== undefined) {
			if (!this.onRemove.has(id)) {
				this.onRemove.set(id, [handler])
				return
			}
			this.onRemove.get(id)!.push(handler)
			return
		}
		id = getParentId(element)
		if (!this.onClear.has(id)) {
			this.onClear.set(id, [handler])
			return
		}
		this.onClear.get(id)!.push(handler)
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
			console.error("Unexpected behavior door [", id, "] not found for the call [" + name + "]")
			return undefined
		}
		return this.getHandler(element._d0r.parent, name)
	}

}

const doors = new Doors()

customElements.define(tag,
	class extends HTMLElement {
		constructor() {
			super()
		}
		connectedCallback() {
			doors.register(this, {})
		}
		disconnectedCallback() {
			doors.unregister(this)
		}
	}
)

export default doors
