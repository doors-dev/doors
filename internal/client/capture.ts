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

import action from "./calls";
import { fetchOpt, fetchOptJson, fetchOptForm, date, result } from "./lib";
import navigator from "./navigator";
import { Fetch, NewFetch } from "./scope";
import { decodePayload } from "./package";
import { HookErr, hookErrKinds } from "./hook_err";
import controller from "./controller";
import { getTrigger } from "./trigger";

// KeyMatch mirrors front.KeyMatch: a key + modifier constraints (0 any, 1 on, 2 off).
interface KeyMatch {
	// key (empty string matches any key)
	k: string;
	// ctrl required state
	c: number;
	// shift required state
	s: number;
	// alt required state
	a: number;
	// meta required state
	m: number;
}

function modOk(pressed: boolean, mod: number): boolean {
	if (mod === 1) {
		return pressed
	}
	if (mod === 2) {
		return !pressed
	}
	return true
}
interface CaptureOpt {
	// preventDefault
	pd?: boolean;
	// stopPropagation
	sp?: boolean;
	// exactTarget
	et?: boolean;
	// keys: key + modifier matches
	ks?: Array<KeyMatch> | null;
	// history.replaceState instead of pushState (link)
	hr?: boolean;
	// exclude value
	ev?: boolean;
}
function applyEventOpt(event: Event, opt: CaptureOpt): boolean {
	if (opt.et) {
		if (event.target !== event.currentTarget) {
			return false
		}
	}
	if (opt.ks && opt.ks.length > 0) {
		const ke = event as KeyboardEvent
		const matched = opt.ks.some((match) => {
			if (match.k !== "" && match.k !== ke.key) {
				return false
			}
			if (!modOk(ke.ctrlKey, match.c)) {
				return false
			}
			if (!modOk(ke.shiftKey, match.s)) {
				return false
			}
			if (!modOk(ke.altKey, match.a)) {
				return false
			}
			if (!modOk(ke.metaKey, match.m)) {
				return false
			}
			return true
		})
		if (!matched) {
			return false
		}
	}
	if (opt.pd) {
		event.preventDefault();
	}
	if (opt.sp) {
		event.stopPropagation();
	}
	return true
}

interface InputValues {
	name: string | null;
	value: string;
	number: number | null;
	date: string | null;
	selected: string[];
	checked: boolean;
}

function getInputValues(input: HTMLInputElement | HTMLSelectElement): InputValues {
	const value = input.value;
	let number: number | null = (input as HTMLInputElement).valueAsNumber;
	if (isNaN(number)) {
		number = null;
	}

	let dateValue: string | null = null;
	const valueAsDate = (input as HTMLInputElement).valueAsDate;
	if (valueAsDate) {
		dateValue = valueAsDate.toISOString();
	}

	let selected: string[] = [];
	if ('selectedOptions' in input && input.selectedOptions) {
		selected = Array.from(input.selectedOptions).map(option => option.value);
	}
	const checked = 'checked' in input ? input.checked === true : false;
	const name = input.name || null;

	return { name, value, number, date: dateValue, selected, checked };
}


export function capture(name: string, opt: any, arg: any, event: Event | undefined, hook: any): Promise<Response> {
	const captureFunction = captures[name]
	if (!captureFunction) {
		throw new HookErr(hookErrKinds.other, new Error("capture " + name + " not found"))
	}
	const [hookId, scopeQueue, indicator, before, , timeout] = hook
	const f = NewFetch({
		hookId,
		event: event,
		scopeQueue,
		indicator,
		before,
		timeout
	})
	return captureFunction(f, arg, opt)
}

type Capture = ((fetch: Fetch, data: any, opt: CaptureOpt) => Promise<Response>)
const captures: { [key: string]: Capture } = {
	"default": (fetch: Fetch, data: any) => {
		return fetch(fetchOpt(data))
	},
	"json": (fetch: Fetch, data: any) => {
		return fetch(fetchOptJson(data))
	},
	"link": async (fetch: Fetch, event: MouseEvent, opt: CaptureOpt) => {
		opt.pd = true;
		if (!applyEventOpt(event, opt)) {
			throw new HookErr(hookErrKinds.canceled)
		}
		const href = (event.currentTarget as HTMLAnchorElement).href!;
		const revert = opt.hr ? navigator.replace(href, false) : navigator.push(href, false)
		if (!revert) {
			throw new HookErr(hookErrKinds.canceled)
		}
		const [res, err] = await result(() => fetch({}))
		if (!err) {
			return res
		}
		if (!(err instanceof HookErr)) {
			throw err
		}
		if (err.network() || err.server()) {
			revert()
			throw err
		}
		if (err.notFound() || err.canceled()) {
			revert()
			throw err
		}
		throw err
	},
	"focus": (fetch: Fetch, event: FocusEvent) => {
		const obj = {
			type: event.type,
			timestamp: date(new Date()),
		};
		return fetch(fetchOptJson(obj))
	},
	"focus_io": (fetch: Fetch, event: FocusEvent, opt: CaptureOpt) => {
		if (!applyEventOpt(event, opt)) {
			throw new HookErr(hookErrKinds.canceled)
		}
		const obj = {
			type: event.type,
			timestamp: date(new Date()),
		};
		return fetch(fetchOptJson(obj))
	},
	"keyboard": (fetch: Fetch, event: KeyboardEvent, opt: CaptureOpt) => {
		if (!applyEventOpt(event, opt)) {
			throw new HookErr(hookErrKinds.canceled)
		}
		const obj = {
			type: event.type,
			key: event.key,
			code: event.code,
			repeat: event.repeat,
			altKey: event.altKey,
			ctrlKey: event.ctrlKey,
			shiftKey: event.shiftKey,
			metaKey: event.metaKey,
			timestamp: date(new Date()),
		};
		return fetch(fetchOptJson(obj))
	},
	"pointer": (fetch: Fetch, event: PointerEvent, opt: CaptureOpt) => {
		if (!applyEventOpt(event, opt)) {
			throw new HookErr(hookErrKinds.canceled)
		}
		const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
		const obj = {
			type: event.type,
			pointerId: event.pointerId,
			pressure: event.pressure,
			tangentialPressure: event.tangentialPressure,
			tiltX: event.tiltX,
			tiltY: event.tiltY,
			twist: event.twist,
			buttons: event.buttons,
			button: event.button,
			pointerType: event.pointerType,
			isPrimary: event.isPrimary,
			pointer: {
				x: event.offsetX,
				y: event.offsetY,
				width: event.width,
				height: event.height,
			},
			target: {
				x: rect.x,
				y: rect.y,
				width: rect.width,
				height: rect.height,
			},
			page: {
				x: window.scrollX,
				y: window.scrollY,
				width: window.innerWidth,
				height: window.innerHeight,
			},
			screen: {
				x: window.screenX,
				y: window.screenY,
				width: window.outerWidth,
				height: window.outerHeight,
			},
			timestamp: date(new Date()),
		};
		return fetch(fetchOptJson(obj))
	},
	"input": (fetch: Fetch, event: InputEvent, opt: CaptureOpt) => {
		return fetch(fetchOptJson({
			type: event.type,
			data: event.data,
			...opt.ev === true ? {} : getInputValues(event.target as HTMLInputElement | HTMLSelectElement),
			timestamp: date(new Date()),
		}))
	},
	"change": (fetch: Fetch, event: Event) => {
		return fetch(fetchOptJson({
			type: event.type,
			...getInputValues(event.target as HTMLInputElement | HTMLSelectElement),
			timestamp: date(new Date()),
		}))
	},
	"submit": (fetch: Fetch, event: SubmitEvent) => {
		applyEventOpt(event, { pd: true });
		const form = event.target as HTMLFormElement;
		const formData = new FormData(form);
		return fetch(fetchOptForm(formData))
	},
}

const attr = "data-d0c"
export function attach(parent: Element | DocumentFragment | Document) {
	for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="applied"])`)) {
		const capturesList = JSON.parse(element.getAttribute(attr)!)
		element.setAttribute(attr, "applied")
		for (const [event, name, opt, hook] of capturesList) {
			const [hookId, scopeQueue, indicator, before, onErr, timeout] = hook
			element.addEventListener(event, async (e) => {
				let p: Promise<Response>
				try {
					p = capture(name, opt, e, e, [hookId, scopeQueue, indicator, before, null, timeout])
				} catch (error: any) {
					p = Promise.reject(error)
				}
				getTrigger(e)?.promises.push(p)
				try {
					await p
				} catch (error: any) {
					if (!(error instanceof HookErr)) {
						console.error("unknown error in capture:", error)
						return
					}
					if (error.canceled() || error.notFound()) {
						return
					}
					if (error.gone()) {
						controller.gone()
						return
					}
					if (!onErr || onErr.length == 0) {
						console.error("capture execution error", error)
						return
					}
					for (const [name, arg, payload] of onErr) {
						const res = action(name, arg, { element, error: error, payload: await decodePayload(payload) })
						if (res instanceof Promise) {
							res.then(([_, e]) => {
								if (e) {
									console.error("error action " + name + " failed", e)
								}
							})
							continue
						}
						const [_, e] = res
						if (e) {
							console.error("error action " + name + " failed", e)
						}
					}
				}
			})
		}
	}
}
