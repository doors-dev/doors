// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import door from './door'
import { capture } from './capture'
import { HookErr, hookErrKinds } from './capture'
import controller from './controller'
import navigator from './navigator'
import { decodePayload } from './package'

function getHookParams(element: HTMLElement, name: string): any | undefined {
	const attrName = `data-d0h-${name}`
	const value = element.getAttribute(attrName)
	if (value == null) return undefined
	return JSON.parse(value)
}


const global: { [key: string]: any } = {}

class $D {
	constructor(private anchor: HTMLElement) { }

	clean = (handler: () => void | Promise<void>): void => {
		door.onUnmount(this.anchor, handler)
	}

	get G() {
		return global
	}

	ready(): Promise<undefined> {
		return controller.ready
	}

	on = (name: string, handler: (arg: any) => any): void => {
		door.on(this.anchor, name, handler)
	}

	fetch = async (name: string, arg: any): Promise<Response> => {
		const hook = getHookParams(this.anchor, name)
		if (hook === undefined) {
			throw new HookErr(hookErrKinds.capture, new Error("hook " + name + " not found"))
		}
		return await capture("default", undefined, arg, undefined, hook)
	}

	hook = async (name: string, arg: any): Promise<any> => {
		const res = await this.fetch(name, arg)
		return await res.json()
	}

	data = async (name: string): Promise<any> => {
		const attrName = `data-d0d-${name}`
		const encodedPayload = this.anchor.getAttribute(attrName)
		if (encodedPayload == null) {
			return undefined
		}
		const payload = await decodePayload(JSON.parse(encodedPayload))
		return payload.any
	}
}


function init(
	anchor: HTMLElement,
	f: (
		$on: $D['on'],
		$data: $D['data'],
		$hook: $D['hook'],
		$fetch: $D['fetch'],
		$G: $D['G'],
		$sys: {
			ready: $D['ready'],
			clean: $D['clean'],
		},
		HookErr: any,
	) => Promise<void> | void
) {
	const $d = new $D(anchor)
	const $on = $d.on
	const $data = $d.data
	const $hook = $d.hook
	const $fetch = $d.fetch
	const $G = $d.G
	const $sys = {
		ready: $d.ready,
		clean: $d.clean,
		activateLinks: () => navigator.activate(document),
	}
	return f($on, $data, $hook, $fetch, $G, $sys, HookErr)
}
export default init
