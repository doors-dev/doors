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
	constructor(private anchor_: HTMLElement) { }

	clean = (handler: () => void | Promise<void>): void => {
		door.onUnmount(this.anchor_, handler)
	}

	get G() {
		return global
	}

	ready(): Promise<void> {
		return controller.ready
	}

	on = (name: string, handler: (arg: any) => any): void => {
		door.on(this.anchor_, name, handler)
	}

	fetch = async (name: string, arg: any): Promise<Response> => {
		const hook = getHookParams(this.anchor_, name)
		if (hook === undefined) {
			throw new HookErr(hookErrKinds.capture, new Error("hook " + name + " not found"))
		}
		return await capture("default", undefined, arg, undefined, hook)
	}

	hook = async (name: string, arg: any): Promise<any> => {
		const res = await this.fetch(name, arg)
		return await res.json()
	}

	data = (name: string): Promise<any> | any => {
		const attrName = `data-d0d-${name}`
		const encodedPayload = this.anchor_.getAttribute(attrName)
		if (encodedPayload == null) {
			return undefined
		}
		const payload = decodePayload(JSON.parse(encodedPayload))
		if (payload instanceof Promise) {
			return (async () => {
				return (await payload).any
			})()
		}
		return payload.any
	}
}


function init(
	anchor: HTMLElement,
	f: (
		$on: (name: string, handler: (arg: any, err?: HookErr) => any) => void,
		$data: $D['data'],
		$hook: (name: string, arg?: any) => Promise<any>,
		$fetch: (name: string, arg?: any) => Promise<Response>,
		$G: $D['G'],
		$sys: {
			ready: $D['ready'],
			clean: $D['clean'],
			activateLinks: () => void,
		},
		HookErr: typeof HookErr,
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
		activateLinks: () => navigator.activateCurrent(),
	}
	return f($on, $data, $hook, $fetch, $G, $sys, HookErr)
}
export default init
