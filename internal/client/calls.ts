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
import navigator from "./navigator"
import indicator, { IndicatorEntry } from "./indicator"
import { removeAttr, setAttr } from "./dyna"
import { doAfter, scrollInto, Result } from "./lib"
import { report } from "./scope.ts"
import { EncodedPayload, Payload } from "./package.ts"
import { HookErr } from "./hook_err.ts"
import { putTrigger } from "./trigger"


type Extras = {
	error?: HookErr,
	payload?: Payload,
	element?: Element,
}

const captureEvents: { [key: string]: (type: string, init: any) => Event } = {
	"pointer": (type, init) => new PointerEvent(type, init),
	"keyboard": (type, init) => new KeyboardEvent(type, init),
	"focus": (type, init) => new FocusEvent(type, init),
	"focus_io": (type, init) => new FocusEvent(type, init),
	"input": (type, init) => new InputEvent(type, init),
	"change": (type, init) => new Event(type, init),
	"submit": (type, init) => new Event(type, init),
}

function syncAttributes(el: Element, attrs: {[key:string]:string}) {
	for (const { name } of Array.from(el.attributes)) {
		if (!(name in attrs)) {
			el.removeAttribute(name)
		}
	}
	for (const [name, value] of Object.entries(attrs)) {
		el.setAttribute(name, value)
	}
}

const actions = {
	"location_reload": (_: Extras) => {
		doAfter(() => {
			location.reload()
		})
	},
	"indicate": (ext: Extras, duration: number, indicatations: Array<IndicatorEntry>) => {
		const id = indicator.start(ext.element ?? null, indicatations)
		if (id) {
			setTimeout(() => indicator.end(id), duration)
		}
	},
	"update_title": (_: Extras, content: string, attrs: {[key:string]:string}) => {
		let title = document.head.querySelector("title")
		if (!title) {
			title = document.createElement("title")
			document.head.appendChild(title)
		}
		title.innerHTML = content
		syncAttributes(title, attrs)
	},
	"remove_meta": (_: Extras, name: string, property: boolean) => {
		const key = property ? "property" : "name"
		let meta = document.head.querySelector(`meta[${key}=${JSON.stringify(name)}]`)
		if(!meta) {
			return
		}
		meta.remove()
	},
	"update_meta": (_: Extras, name: string, property: boolean, attrs: {[key:string]:string}) => {
		const key = property ? "property" : "name"
		const targetAttrs = {
			...attrs,
			[key]: name,
		}
		let meta = document.head.querySelector(`meta[${key}=${JSON.stringify(name)}]`)
		if (!meta) {
			meta = document.createElement("meta")
			document.head.appendChild(meta)
		}
		syncAttributes(meta, targetAttrs)
	},
	"report_hook": (_: Extras, track: number) => {
		report(track)
	},
	"emit_event": (ext: Extras, emitterId: number, type: string, capture: string) => {
		const promises: Promise<Response>[] = []
		const elements = doors.getEmitter(emitterId)
		if (elements) {
			const make = captureEvents[capture]
			for (const element of elements) {
				let event: Event
				if (make) {
					event = make(type, { bubbles: true, cancelable: true, ...ext.payload?.any })
				} else {
					event = new CustomEvent(type, { bubbles: true, cancelable: true, detail: ext.payload?.any })
				}
				const meta = putTrigger(event)
				element.dispatchEvent(event)
				promises.push(...meta.promises)
			}
		}
		return Promise.all(promises).then(
			() => promises.length,
			(err) => {
				if (err instanceof HookErr) {
					throw new Error(err.message ? `${err.kind}: ${err.message}` : err.kind)
				}
				throw err
			},
		)
	},
	"location_replace": (_: Extras, href: string, origin: boolean) => {
		let url: URL
		if (origin) {
			url = new URL(href, window.location.origin);
		} else {
			url = new URL(href)
		}
		doAfter(() => {
			location.replace(url.toString())
		})
	},
	"scroll": (_: Extras, selector: string, options: any) => {
		if(!scrollInto(selector, options)) {
			throw new Error("element to scroll into not found")
		}
	},
	"location_assign": (_: Extras, href: string, origin: boolean) => {
		let url: URL
		if (origin) {
			url = new URL(href, window.location.origin);
		} else {
			url = new URL(href)
		}
		doAfter(() => {
			location.assign(url.toString())
		})
	},
	"emit": (ext: Extras, name: string, doorId: number): any => {
		const handler = doors.getHandler(doorId, name)
		if (!handler) {
			throw new Error(`handler ${name} not found`)
		}
		return handler(ext.payload!.any, ext.error as any)
	},
	"dyna_set": (_: Extras, id: number, value: string) => {
		setAttr(id, value)
	},
	"dyna_remove": (_: Extras, id: number) => {
		removeAttr(id)
	},
	"set_path": (_: Extras, path: string, replace: boolean) => {
		if (replace) {
			navigator.replace(path, true)
			return
		}
		navigator.push(path, true)
	},
	"door_replace": (ext: Extras, doorId: number) => {
		doors.replace(doorId, ext.payload!.text!)
	},
	"door_update": (ext: Extras, doorId: number) => {
		doors.update(doorId, ext.payload!.text!)
	},
}

type Err = {
	message: string;
	[key: string]: any;
};

export type Action = [string, Array<any>, EncodedPayload] | [string, Array<any>]

function normalizeErr(e: any): Err {
	if (e && typeof e === "object" && typeof e.message === "string") {
		return e
	}
	return new Error("unknown error")
}

export default function action(name: string, args: Array<any>, extras: Extras = {}): Result<any, Err> | Promise<Result<any, Err>> {
	try {
		const fn = actions[name]
		if (!fn) {
			throw new Error(`action [${name}] not found`)
		}
		let output = fn(extras, ...args)
		if (output instanceof Promise) {
			return output.then(
				(o): Result<any, Err> => [o ?? null, undefined],
				(e): Result<any, Err> => [undefined, normalizeErr(e)],
			)
		}
		if (output === undefined) {
			output = null
		}
		return [output, undefined]
	} catch (e) {
		return [undefined, normalizeErr(e)]
	}
}
