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

import action from "./calls"

import doors from "./door"
import navigator from "./navigator"
import { Package } from "./package";
import { Connector, Frame, NewConnector } from "./connector";
import { Solitaire } from "./deck";
import { Tracker } from "./tracker";
import { disconnectAfter, id, ttl, requestTimeout, prefix, solitairePending } from "./params";
import { ReliableTimer, ProgressiveDelay, AbortTimer, result } from "./lib";


const state = {
	dead: "dead",
	sleep: "sleep",
	active: "active",
} as const

type State = typeof state[keyof typeof state]


function repeatedLoad(): boolean {
	if (history.state?._d0rI === id) {
		return true
	}
	history.replaceState({ ...history.state, _d0rI: id }, '')
	return false
}

class Controller {
	private state: State = state.active
	private deck = new Solitaire()
	private tracker = new Tracker(() => this.send(), p => action(p.action, p.arg, { payload: p.getPayload() }), solitairePending)
	readonly ready: Promise<void>
	private connector: Connector
	private loaded = false
	private ttlTimer = new ReliableTimer(ttl, () => this.suspend())
	constructor() {
		let ready: any = undefined
		this.ready = new Promise((res) => { ready = res })
		if (repeatedLoad()) {
			this.ensureReload()
			return
		}
		this.pushState()
		document.addEventListener("DOMContentLoaded", () => {
			this.init()
			ready()
			if (this.state == state.dead) {
				return
			}
			this.flush()
		})
		this.connector = NewConnector(this)
		this.syncVisibility()
		window.addEventListener("pagehide", () => this.sleep())
		window.addEventListener("pageshow", () => this.syncVisibility())
		document.addEventListener("visibilitychange", () => this.syncVisibility())
	}
	private async pushState() {
		const delay = new ProgressiveDelay()
		const body = JSON.stringify(navigator.getState())
		while (this.state != state.dead) {
			const abortTimer = new AbortTimer(requestTimeout)
			const [response, err] = await result(() => fetch(`${prefix}/t/${id}`, {
				method: "POST",
				signal: abortTimer.signal,
				headers: {
					'Content-Type': 'application/json;charset=UTF-8',
				},
				body,
			}))
			abortTimer.cancel()
			if (!err && response.ok) {
				return
			}
			if (!err && response.status === 410) {
				this.kill()
				return
			}
			await delay.wait()
		}
	}
	private ensureReload() {
		this.state = state.dead
		this.reload()
		this.reloadOnTouch()
		window.addEventListener("pageshow", () => this.reload())
	}
	resetDelays() {
		this.connector.resetDelays()
	}
	resetTTL() {
		this.ttlTimer.reset()
	}
	private init() {
		this.loaded = true
		doors.scan(document)
	}
	private sleepTimer: number | undefined = undefined
	private syncVisibility() {
		if (this.state == state.dead) {
			return
		}
		if (document.hidden && this.state == state.sleep) {
			return
		}
		if (document.hidden && this.sleepTimer !== undefined) {
			return
		}
		if (document.hidden) {
			this.sleepTimer = setTimeout(() => {
				this.sleep()
			}, disconnectAfter)
			return
		}
		clearTimeout(this.sleepTimer)
		this.sleepTimer = undefined
		if (this.state == state.sleep) {
			this.state = state.active
			this.connector.resume()
			return
		}
	}
	private sleep() {
		clearTimeout(this.sleepTimer)
		this.sleepTimer = undefined
		if (this.state != state.active) {
			return
		}
		this.state = state.sleep
		this.connector.pause()
	}
	ack(frame: Frame) {
		this.deck.return(frame.lost)
	}
	drop(frame: Frame) {
		this.tracker.buffer(frame.results)
		this.deck.return(frame.lost)
		this.send()
	}
	onPackage(p: Package) {
		this.deck.insert(p)
		if (!this.loaded) {
			return
		}
		this.flush()
	}
	private flush() {
		const collection = this.deck.collect()
		for (const p of collection) {
			this.tracker.push(p)
		}
		this.send()
	}
	private send() {
		const results = this.tracker.collect()
		const lost = this.deck.collectLost()
		if (lost.length == 0 && results.size == 0) {
			return
		}
		this.connector.send({ results, lost })
	}
	suspend() {
		if (this.state == state.dead) {
			return
		}
		this.state = state.dead;
		this.connector.pause()
		this.reloadOnTouch()
	}
	kill() {
		if (this.state == state.dead) {
			if (!document.hidden) {
				this.reload()
			}
			return
		}
		this.state = state.dead
		this.connector.pause()
		if (!document.hidden) {
			this.reload()
		}
		this.reloadOnTouch()
	}
	private reloadOnTouch() {
		["pointerdown", "pointermove", "pointerup", "scroll", "focus", "keydown", "input"].forEach(event => {
			window.addEventListener(event, () => this.reload(), true);
		});
	}
	private reloaded = false
	private reload() {
		if (this.reloaded) {
			return
		}
		this.reloaded = true
		location.reload()
		new ReliableTimer(requestTimeout, () => {
			this.reloaded = false
		})
	}
}

const controller = new Controller()

export default {
	ready: controller.ready,
	resetDelays() {
		controller.resetDelays()
	},
	gone() {
		controller.kill()
	}
}
