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
import { Connector, Frame, Lost, NewConnector, Results, STRESS_MODE } from "./connector";
import { disconnectAfter, id, ttl, requestTimeout, prefix, solitairePending } from "./params";
import { ReliableTimer, ProgressiveDelay, AbortTimer, result } from "./lib";


class Solitaire {
	top: Card
	private cursor_: number = 1
	constructor() {

	}
	private collectedLost_ = new Set<number>()
	return(lost: Lost) {
		for (const gap of lost) {
			this.collectedLost_.delete(gap)
		}
	}
	collectLost(): Lost {
		const lost: Lost = []
		for (const seq of this.getLost()) {
			if (this.collectedLost_.has(seq)) {
				continue
			}
			lost.push(seq)
			this.collectedLost_.add(seq)
		}
		if (STRESS_MODE && lost.length != 0) {
			console.log("new lost:", lost)
		}
		return lost

	}
	private getLost(): Lost {
		let lost: Lost = []
		if (!this.top) {
			return lost
		}
		this.top.lost(this.cursor_ - 1, lost)
		return lost
	}
	collect(): Array<Package> {
		const a: Array<Package> = []
		if (!this.top) {
			return a
		}
		if (this.top.start > this.cursor_) {
			return a
		}
		this.top.collect(a, this)
		const tail = a[a.length - 1]
		this.cursor_ = tail.end + 1
		return a.filter(p => !p.isFiller)
	}
	insert(p: Package) {
		if (p.end < this.cursor_) {
			return
		}
		if (!this.top) {
			this.top = new Card(p)
			return
		}
		this.top.insert(p)
	}

}

class Card {
	private next_: Card
	constructor(private p_: Package) {
	}
	get end(): number {
		return this.p_.end
	}
	get start(): number {
		return this.p_.start
	}
	lost(end: number, lost: Lost) {
		if (end < this.p_.start - 1) {
			const lostStart = end + 1
			const lostEnd = this.p_.start - 1
			for (let seq = lostStart; seq <= lostEnd; seq++) {
				lost.push(seq)
			}
		}
		if (this.next_) {
			this.next_.lost(this.p_.end, lost)
		}
	}
	collect(a: Array<Package>, h: Solitaire) {
		let match = false
		if (a.length == 0) {
			match = true
		} else {
			const tail = a[a.length - 1]
			match = tail.end == this.start - 1
		}
		if (!match) {
			return
		}
		a.push(this.p_)
		h.top = this.next_
		if (this.next_) {
			this.next_.collect(a, h)
		}
	}
	insert(p: Package) {
		if (p.end < this.p_.start) {
			if (this.next_) {
				this.next_.insert(this.p_)
			} else {
				this.next_ = new Card(this.p_)
			}
			this.p_ = p
			return
		}
		if (p.start > this.p_.end) {
			if (this.next_) {
				this.next_.insert(p)
			} else {
				this.next_ = new Card(p)
			}
			return
		}
		if (p.start == this.p_.end) {
			p.start = this.p_.start
			this.p_ = p
			if (this.next_) {
				this.next_.cover(this.p_.end, this)
			}
			return
		}
		// start < end
		if (p.start <= this.p_.start && p.end >= this.p_.end) {
			this.p_ = p
			if (this.next_) {
				this.next_.cover(this.p_.end, this)
			}
			return
		}
		if (p.end >= this.p_.end) { // && p.start > this.start
			p.start = this.p_.start
			this.p_ = p
			if (this.next_) {
				this.next_.cover(this.p_.end, this)
			}
			return
		}
		// p.end < this.end
		p.start = Math.min(this.p_.start, p.start)
		this.p_.start = p.end + 1
		if (this.next_) {
			this.next_.insert(this.p_)
		} else {
			this.next_ = new Card(this.p_)
		}
		this.p_ = p
	}
	private cover(end: number, head: Card) {
		if (end >= this.p_.end) {
			head.next_ = this.next_
			if (this.next_) {
				this.next_.cover(end, head)
			}
			return
		}
		if (this.p_.start > end) {
			return
		}
		this.p_.start = end + 1
		return
	}
}


class Batch {
	private packages: Array<Package> = []
	private parents: Set<Batch> = new Set()
	children: Set<Batch> = new Set()
	private finished = false
	id: number = 0
	tag: number = 0
	constructor(private tracker: Tracker, init: { id?: number, tag?: number, finished?: boolean }) {
		this.finished = init.finished === true
		this.id = init.id ?? 0
		this.tag = init.tag ?? 0
	}
	empty(): boolean {
		return this.packages.length == 0
	}
	finish(closer: Package) {
		this.finished = true
		this.packages.push(closer)
	}
	ready(seen: Set<Batch>, force?: boolean): boolean {
		if (seen.has(this)) {
			return true
		}
		if (force) {
			this.finished = true
		}
		if (!this.finished) {
			return false
		}
		seen.add(this)
		for (const b of this.parents) {
			if (!b.ready(seen, force)) {
				return false
			}
		}
		return true
	}
	release(): Array<Package> {
		this.tracker.removeBatch(this)
		for (const p of this.packages) {
			this.tracker.unhold(p, this)
		}
		for (const b of this.parents) {
			b.children.delete(this)
		}
		for (const b of this.children) {
			b.parents.delete(this)
		}
		return this.packages
	}
	add(p: Package) {
		this.packages.push(p)
		this.wait(this.tracker.doorHolder(p.door))
		this.wait(this.tracker.packageHolder(p.dependency))
		this.wait(this.tracker.doorHolder(p.dependencyDoor))
		this.tracker.hold(p, this)
	}
	waitAll(batches: Set<Batch> | undefined) {
		if (batches == undefined) {
			return
		}
		for (const batch of batches) {
			this.wait(batch)
		}
	}
	private wait(batch: Batch | undefined) {
		if (batch == undefined || batch == this) {
			return
		}
		this.parents.add(batch)
		batch.children.add(this)
	}
}

class Tracker {
	private buffered_: Results = new Map()
	constructor(private notify_: () => void) {
	}
	private batches = new Set<Batch>()
	private idLookup = new Map<number, Batch>()
	private doorLookup = new Map<number, Batch>()
	private packageLookup = new Map<number, Batch>()
	private tagLookup = new Map<number, Set<Batch>>()
	removeBatch(batch: Batch) {
		this.batches.delete(batch)
		this.idLookup.delete(batch.id)
		const tagged = this.tagLookup.get(batch.tag)
		if (tagged == undefined) {
			return
		}
		tagged.delete(batch)
		if (tagged.size == 0) {
			this.tagLookup.delete(batch.tag)
		}
	}
	doorHolder(door: number): (Batch | undefined) {
		return this.doorLookup.get(door)
	}
	packageHolder(seq: number): (Batch | undefined) {
		return this.packageLookup.get(seq)
	}
	hold(p: Package, batch: Batch) {
		this.doorLookup.set(p.door, batch)
		this.packageLookup.set(p.end, batch)
	}
	unhold(p: Package, batch: Batch) {
		this.packageLookup.delete(p.end)
		if (this.doorLookup.get(p.door) == batch) {
			this.doorLookup.delete(p.door)
		}
	}
	push(p: Package) {
		this.process(p)
		while (this.packageLookup.size >= solitairePending) {
			for (const batch of this.batches) {
				if (batch.empty()) {
					continue
				}
				this.flush(batch, true)
				break
			}
		}
	}
	private process(p: Package) {
		if (p.action == "batch_init") {
			if (this.idLookup.has(p.batch)) {
				this.buffered_.set(p.end, [undefined, "batch is already intialized"])
				return
			}
			const batch = new Batch(this, { id: p.batch })
			this.batches.add(batch)
			this.idLookup.set(p.batch, batch)
			this.execute(p)
			return
		}
		const closer = p.batch > 0 && (p.action == "batch_flush" || p.action == "report_hook")
		const hard = this.idLookup.get(p.batch)
		if (hard && closer) {
			this.idLookup.delete(p.batch)
			hard.finish(p)
			this.flush(hard)
			return
		}
		if (hard) {
			hard.add(p)
			return
		}
		const batch = new Batch(this, { finished: true, tag: closer ? 0 : p.batch })
		this.batches.add(batch)
		batch.add(p)
		if (closer) {
			batch.waitAll(this.tagLookup.get(p.batch))
		}
		this.flush(batch)
		if (batch.tag == 0 || !this.batches.has(batch)) {
			return
		}
		let tagged = this.tagLookup.get(batch.tag)
		if (tagged == undefined) {
			tagged = new Set()
			this.tagLookup.set(batch.tag, tagged)
		}
		tagged.add(batch)
	}
	flush(batch: Batch, force?: boolean) {
		const ready = new Set<Batch>()
		const queue = [batch]
		for (const b of queue) {
			const seen = new Set(ready)
			if (!b.ready(seen, force && b == batch)) {
				continue
			}
			for (const r of seen) {
				if (ready.has(r)) {
					continue
				}
				ready.add(r)
				queue.push(...r.children)
			}
		}
		const packages: Array<Package> = []
		for (const b of ready) {
			packages.push(...b.release())
		}
		packages.sort((a, b) => a.end - b.end)
		for (const p of packages) {
			this.execute(p)
		}
	}
	execute(p: Package) {
		const result = action(p.action, p.arg, { payload: p.getPayload() })
		if (result instanceof Promise) {
			result.then(([ok, err]) => {
				this.buffered_.set(p.end, [ok, err?.message])
				this.notify_()
			})
			return
		}
		const [ok, err] = result
		this.buffered_.set(p.end, [ok, err?.message])
	}
	buffer(collected: Results) {
		for (const [seq, entry] of collected.entries()) {
			this.buffered_.set(seq, entry)
		}
	}
	collect(): Results {
		const collected = this.buffered_
		this.buffered_ = new Map()
		return collected
	}
}

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
	private tracker_ = new Tracker(() => this.send())
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
