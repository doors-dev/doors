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
	private state_: State = state.active
	private deck_ = new Solitaire()
	private tracker_ = new Tracker(() => this.send(), p => action(p.action, p.arg, { payload: p.getPayload() }), solitairePending)
	readonly ready: Promise<void>
	private connector_: Connector
	private loaded_ = false
	private ttlTimer_ = new ReliableTimer(ttl, () => this.suspend())
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
			if (this.state_ == state.dead) {
				return
			}
			this.flush()
		})
		this.connector_ = NewConnector(this)
		this.syncVisibility()
		window.addEventListener("pagehide", () => this.sleep())
		window.addEventListener("pageshow", () => this.syncVisibility())
		document.addEventListener("visibilitychange", () => this.syncVisibility())
	}
	private async pushState() {
		const delay = new ProgressiveDelay()
		const body = JSON.stringify(navigator.getState())
		while (this.state_ != state.dead) {
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
		this.state_ = state.dead
		this.reload()
		this.reloadOnTouch()
		window.addEventListener("pageshow", () => this.reload())
	}
	resetDelays() {
		this.connector_.resetDelays()
	}
	resetTTL() {
		this.ttlTimer_.reset()
	}
	private init() {
		this.loaded_ = true
		doors.scan(document)
	}
	private sleepTimer_: number | undefined = undefined
	private syncVisibility() {
		if (this.state_ == state.dead) {
			return
		}
		if (document.hidden && this.state_ == state.sleep) {
			return
		}
		if (document.hidden && this.sleepTimer_ !== undefined) {
			return
		}
		if (document.hidden) {
			this.sleepTimer_ = setTimeout(() => {
				this.sleep()
			}, disconnectAfter)
			return
		}
		clearTimeout(this.sleepTimer_)
		this.sleepTimer_ = undefined
		if (this.state_ == state.sleep) {
			this.state_ = state.active
			this.connector_.resume()
			return
		}
	}
	private sleep() {
		clearTimeout(this.sleepTimer_)
		this.sleepTimer_ = undefined
		if (this.state_ != state.active) {
			return
		}
		this.state_ = state.sleep
		this.connector_.pause()
	}
	ack(frame: Frame) {
		this.deck_.return(frame.lost)
	}
	drop(frame: Frame) {
		this.tracker_.buffer(frame.results)
		this.deck_.return(frame.lost)
		this.send()
	}
	onPackage(p: Package) {
		this.deck_.insert(p)
		if (!this.loaded_) {
			return
		}
		this.flush()
	}
	private flush() {
		const collection = this.deck_.collect()
		for (const p of collection) {
			this.tracker_.push(p)
		}
		this.send()
	}
	private send() {
		const results = this.tracker_.collect()
		const lost = this.deck_.collectLost()
		if (lost.length == 0 && results.size == 0) {
			return
		}
		this.connector_.send({ results, lost })
	}
	suspend() {
		if (this.state_ == state.dead) {
			return
		}
		this.state_ = state.dead;
		this.connector_.pause()
		this.reloadOnTouch()
	}
	kill() {
		if (this.state_ == state.dead) {
			if (!document.hidden) {
				this.reload()
			}
			return
		}
		this.state_ = state.dead
		this.connector_.pause()
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
	private reloaded_ = false
	private reload() {
		if (this.reloaded_) {
			return
		}
		this.reloaded_ = true
		location.reload()
		new ReliableTimer(requestTimeout, () => {
			this.reloaded_ = false
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
