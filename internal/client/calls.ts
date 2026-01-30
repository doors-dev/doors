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
import { doAfter } from "./lib"
import { report } from "./scope.ts"


type Extras = {
	error?: HookErr,
	payload?: ArrayBuffer,
	element?: Element,
}

const decoder = new TextDecoder()

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
	scroll: (_: Extras, selector: string, smooth: boolean) => {
		const el = document.querySelector(selector)
		if (el) {
			el.scrollIntoView({ behavior: smooth ? "smooth" : "auto" });
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
	emit: (ext: Extras, name: string, arg: any, doorId: number): any => {
		const handler = doors.getHandler(doorId, name)
		if (!handler) {
			throw new Error(`Handler ${name} not found`)
		}
		return handler(arg, ext.error as any)
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
		navigator.push(path)
	},
	door_replace: (ext: Extras, doorId: number) => {
		doors.replace(doorId, decoder.decode(ext.payload!))
	},
	door_update: (ext: Extras, doorId: number) => {
		doors.update(doorId, decoder.decode(ext.payload!))
	},
}

type Output = Exclude<any, undefined>;
type Err = {
	message: string;
	[key: string]: any;
};

export type CallResult = ([Output, undefined] | [undefined, Err])

export type Action = [string, Array<any>]

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
		if (e && typeof e === "object" && typeof e.message === "string") {
			return [undefined, e]
		}
		return [undefined, new Error("unknown error")]
	}
}
