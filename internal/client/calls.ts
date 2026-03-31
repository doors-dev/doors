// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import doors from "./door"
import navigator from "./navigator"
import indicator, { IndicatorEntry } from "./indicator"
import { removeAttr, setAttr } from "./dyna"
import { HookErr } from "./capture"
import { doAfter, scrollInto } from "./lib"
import { report } from "./scope.ts"
import { EncodedPayload, Payload } from "./package.ts"


type Extras = {
	error?: HookErr,
	payload?: Payload,
	element?: Element,
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
	location_reload: (_: Extras) => {
		doAfter(() => {
			location.reload()
		})
	},
	indicate: (ext: Extras, duration: number, indicatations: Array<IndicatorEntry>) => {
		const id = indicator.start(ext.element ?? null, indicatations)
		if (id) {
			setTimeout(() => indicator.end(id), duration)
		}
	},
	update_title: (_: Extras, content: string, attrs: {[key:string]:string}) => {
		let title = document.head.querySelector("title")
		if (!title) {
			title = document.createElement("title")
			document.head.appendChild(title)
		}
		title.textContent = content
		syncAttributes(title, attrs)
	},
	update_meta: (_: Extras, name: string, property: boolean, attrs: {[key:string]:string}) => {
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
	report_hook: (_: Extras, track: number) => {
		report(track)
	},
	location_replace: (_: Extras, href: string, origin: boolean) => {
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
	scroll: (_: Extras, selector: string, options: any) => {
		if(!scrollInto(selector, options)) {
			throw new Error("element to scroll into not found")
		}
	},
	location_assign: (_: Extras, href: string, origin: boolean) => {
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
	emit: (ext: Extras, name: string, doorId: number): any => {
		const handler = doors.getHandler(doorId, name)
		if (!handler) {
			throw new Error(`Handler ${name} not found`)
		}
		return handler(ext.payload!.any, ext.error as any)
	},
	dyna_set: (_: Extras, id: number, value: string) => {
		setAttr(id, value)
	},
	dyna_remove: (_: Extras, id: number) => {
		removeAttr(id)
	},
	set_path: (_: Extras, path: string, replace: boolean) => {
		if (replace) {
			navigator.replace(path)
			return
		}
		navigator.push(path, true)
	},
	door_replace: (ext: Extras, doorId: number) => {
		doors.replace(doorId, ext.payload!.text!)
	},
	door_update: (ext: Extras, doorId: number) => {
		doors.update(doorId, ext.payload!.text!)
	},
}

type Output = Exclude<any, undefined>;
type Err = {
	message: string;
	[key: string]: any;
};

export type CallResult = ([Output, undefined] | [undefined, Err])

export type Action = [string, Array<any>, EncodedPayload] | [string, Array<any>]

export default function action(name: string, args: Array<any>, extras: Extras = {}): CallResult {
	try {
		const fn = actions[name]
		if (!fn) {
			throw new Error(`action [${name}] not found`)
		}
		let output = fn(extras, ...args)
		if (output instanceof Promise) {
			throw new Error("async actions are prohibited")
		}
		if (output === undefined) {
			output = null
		}
		return [output, undefined]
	} catch (e) {
		console.error(e)
		if (e && typeof e === "object" && typeof e.message === "string") {
			return [undefined, e]
		}
		return [undefined, new Error("unknown error")]
	}
}
