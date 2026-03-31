// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import { id, disconnectAfter, ttl, solitairePing, prefix } from "./params"
import action from "./calls"
import { ProgressiveDelay, AbortTimer, ReliableTimer } from "./lib"

import doors from "./door"
import { Package, Header } from "./package";

const STRESS_MODE = false

const controlBytes = {
	terminator: 0xFF,
	discard: 0xFD,
}
const signals = {
	ack: 0x00,
	action: 0x01,
	roll: 0x02,
	suspend: 0x03,
	kill: 0x04,
}


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
	isDone(): boolean {
		return !this.top
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
		if (tail.end < this.cursor_) {
			debugger;
		}
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

const connectorStatus = {
	signal: "signal",
	header: "header",
	payload: "payload",
} as const
type SyncStatus = typeof connectorStatus[keyof typeof connectorStatus];


class NetworkError extends Error { }

const reports = {
	ok: "ok",
	broken: "broken",
	interrupted: "interrupt",
} as const;
type Report = typeof reports[keyof typeof reports];

type Lost = Array<number>
type Gap = ([number, number] | [number])
type Gaps = Array<Gap>


class Connection {
	private status_: SyncStatus = connectorStatus.signal
	private abortTimer_: AbortTimer
	private rollTimer_: ReliableTimer
	constructor(private ctrl_: Controller, private id_: number, private results_: Results, private lost_: Lost) {
		this.abortTimer_ = new AbortTimer(solitairePing * 4 / 3)
		this.rollTimer_ = new ReliableTimer(solitairePing, () => {
			this.ctrl_.requestRoll(this)
		})
		this.run()
	}
	private header_: Header | undefined
	private package_: Package | undefined

	abort() {
		this.abortTimer_.abort()
	}
	private acked_ = false
	private ack() {
		if (this.acked_) {
			return
		}
		this.acked_ = true

	}
	private report(ok: boolean = false) {
		this.abortTimer_.cancel()
		this.rollTimer_.cancel()
		const report = ok ? reports.ok : this.acked_ ? reports.interrupted : reports.broken;
		this.ctrl_.report(this, this.id_, report, this.results_, this.lost_)
	}
	private get gaps(): Gaps {
		const gaps: Gaps = []
		let gap: any
		for (const seq of this.lost_) {
			if (!gap) {
				gap = [seq]
				continue
			}
			const prev = gap[1] !== undefined ? gap[1] : gap[0]
			if (seq == prev + 1) {
				gap[1] = seq
				continue
			}
			gaps.push(gap)
			gap = [seq]
		}
		if (gap) {
			gaps.push(gap)
		}
		return gaps
	}

	private async run() {
		try {
			let response: Response
			try {
				response = await fetch(`${prefix}/s/${id}`, {
					signal: this.abortTimer_.signal,
					method: "PUT",
					headers: {
						Accept: "application/octet-stream",
						'Content-Type': 'application/json;charset=UTF-8',
					},
					body: JSON.stringify({
						gaps: this.gaps,
						results: Object.fromEntries(this.results_!),
					}),
				})
			} catch (e) {
				throw new NetworkError()
			}
			if (response.status === 401 || response.status === 410) {
				this.ctrl_.kill()
				throw new Error()
			}
			if (!response.ok) {
				throw new Error()
			}
			const reader = response.body!.getReader()
			while (true) {
				let value: Uint8Array
				const result = await reader.read()
				if (result.done) {
					throw new Error()
				}
				if (STRESS_MODE && Math.random() > 0.5) {
					throw new Error()
				}
				value = result.value
				const done = await this.onChunk(value)
				this.ctrl_.flush()
				if (done) {
					reader.cancel()
					break
				}
			}
			this.report(true)
		} catch (e) {
			this.report()
		}
	}
	private async onChunk(data: Uint8Array): Promise<boolean> {
		if (data.length == 0) {
			return false
		}
		if (this.status_ == connectorStatus.signal) {
			const signal = data[0]
			switch (signal) {
				case signals.ack:
					this.ack()
					if (data.length == 1) {
						return false
					}
					return await this.onChunk(data.subarray(1))
				case signals.action:
					this.status_ = connectorStatus.header
					this.header_ = new Header()
					if (data.length == 1) {
						return false
					}
					return await this.onChunk(data.subarray(1))
				case signals.suspend:
					this.ctrl_.suspend()
					break;
				case signals.kill:
					this.ctrl_.kill()
					break;
				case signals.roll:
					break
				default:
					console.error(new Error("unsupported signal " + signal))
			}
			return true
		}
		if (this.status_ == connectorStatus.header) {
			for (let i = 0; i < data.length; i++) {
				const byte = data[i]
				if (byte == controlBytes.terminator) {
					this.header_!.append(data.subarray(0, i))
					this.package_ = await this.header_!.package()
					this.header_ = undefined
					this.status_ = connectorStatus.payload
					if (await this.package_!.finalize()) {
						this.ctrl_.onPackage(this.id_, this.package_)
						this.package_ = undefined
						this.status_ = connectorStatus.signal
					}
				} else if (byte == controlBytes.discard) {
					this.header_ = undefined
					this.status_ = connectorStatus.signal
				} else {
					continue
				}
				if (i + 1 == data.length) {
					return false
				}
				return await this.onChunk(data.subarray(i + 1))
			}
			this.header_!.append(data)
			return false
		}
		const remaining = this.package_!.remaining();
		const chunk = remaining >= data.length ? data : data.subarray(0, remaining)
		this.package_!.append(chunk)
		if (await this.package_!.finalize()) {

			if (STRESS_MODE && Math.random() > 0.5) {
				const p = this.package_!
				setTimeout(() => {
					this.ctrl_.onPackage(this.id_, p)
					this.ctrl_.flush()
				}, Math.round(Math.random() * 200))
			} else {
				this.ctrl_.onPackage(this.id_, this.package_!)
			}
			this.package_ = undefined
			this.status_ = connectorStatus.signal
		}
		if (chunk.length == data.length) {
			return false
		}
		return this.onChunk(data.subarray(chunk.length))
	}
}

type Results = Map<number, [any, undefined] | [undefined, string]>

class Tracker {
	private buffered_: Results = new Map()
	process(p: Package) {
		const [ok, err] = action(p.action, p.arg, { payload: p.getPayload() })
		this.buffered_.set(p.end, [ok, err?.message])
	}
	return(collected: Results) {
		for (const [seq, entry] of collected.entries()) {
			this.buffered_.set(seq, entry)
		}
	}
	collect(): Results {
		const collected = this.buffered_
		this.buffered_ = new Map()
		return collected
	}
	isDone(): boolean {
		return this.buffered_.size == 0
	}
}

const state = {
	dead: "dead",
	sleep: "sleep",
	active: "active",
} as const

type State = typeof state[keyof typeof state]

class Controller {
	private connections_ = new Set<Connection>()
	private state_: State = state.active
	private loaded_ = false
	private delay_ = new ProgressiveDelay()
	private counter_ = 0
	deck = new Solitaire()
	tracker = new Tracker()
	ready: Promise<undefined>
	constructor() {
		let ready: any
		this.ready = new Promise((res) => {
			ready = res
		})
		window.addEventListener("pagehide", () => {
			this.sleep()
		})
		document.addEventListener("visibilitychange", () => {
			if (!document.hidden) {
				this.showed()
			} else {
				this.hidden()
			}
		})
		document.addEventListener("DOMContentLoaded", () => {
			this.onReady()
			ready()
		})
		this.roll()
	}
	private ttlTimer = new ReliableTimer(ttl, () => this.suspend())
	private resetTtl() {
		this.ttlTimer.reset()
	}
	private onReady() {
		this.loaded_ = true
		doors.scan(document)
		if (this.state_ == state.dead) {
			return
		}
		this.flush()
	}
	flush() {
		if (!this.loaded_) {
			return
		}
		const collection = this.deck.collect()
		for (const p of collection) {
			this.tracker.process(p)
		}
		this.resetTtl()
		if (this.deck.isDone() && this.tracker.isDone()) {
			return
		}
		this.roll()
	}
	private rolling_ = false
	private roll(delay = false) {
		if (this.state_ != state.active) {
			return
		}
		if (this.rolling_) {
			return
		}
		if (!STRESS_MODE && delay) {
			this.rolling_ = true
			this.delay_.wait().then(() => {
				this.rolling_ = false
				this.connect_()
			})
			return
		}
		this.connect_()
	}
	private connect_() {
		const results = this.tracker.collect()
		const lost = this.deck.collectLost()
		if (lost.length == 0 && results.size == 0 && this.connections_.size > 1) {
			return
		}
		this.counter_ += 1
		this.connections_.add(new Connection(this, this.counter_, results, lost))
	}
	requestRoll(connection: Connection) {
		if (!this.connections_.has(connection)) {
			return
		}
		if (this.connections_.size != 1) {
			return
		}
		this.roll()
	}
	private boundary_ = 1
	private boundaryBuffer_ = new Set<number>
	private packageBuffer_: Array<{ id: number, pkg: Package }> = []
	onPackage(id: number, pkg: Package) {
		if (id <= this.boundary_) {
			this.deck.insert(pkg)
			return
		}
		this.packageBuffer_.push({ id, pkg })
	}
	private updateBoundary(id: number) {
		this.boundaryBuffer_.add(id)
		if (this.boundary_ != id) {
			return
		}
		while (this.boundaryBuffer_.has(this.boundary_)) {
			this.boundaryBuffer_.delete(this.boundary_)
			this.boundary_ += 1
		}
		const packageBuffer: Array<{ id: number, pkg: Package }> = []
		for (const pb of this.packageBuffer_) {
			if (pb.id <= this.boundary_) {
				this.deck.insert(pb.pkg)
				continue
			}
			packageBuffer.push(pb)
		}
		this.packageBuffer_ = packageBuffer
	}
	report(connection: Connection, id: number, report: Report, results: Results, lost: Lost) {
		this.connections_.delete(connection)
		this.deck.return(lost)
		if (report == reports.broken) {
			this.tracker.return(results)
		}
		this.updateBoundary(id)
		if (report == reports.ok) {
			this.delay_.reset()
		}
		if (this.connections_.size >= 6 || (this.connections_.size != 0 && report == reports.ok)) {
			return
		}
		this.roll(report == reports.broken)
	}


	private sleepTimer_: any = null
	private sleep() {
		clearTimeout(this.sleepTimer_)
		this.sleepTimer_ = null
		if (this.state_ != state.active) {
			return
		}
		this.state_ = state.sleep
		this.closeConnections()
	}
	private hidden() {
		if (this.state_ != state.active) {
			return
		}
		if (this.sleepTimer_ != null) {
			return
		}
		this.sleepTimer_ = setTimeout(() => {
			this.sleep()
		}, disconnectAfter)
	}

	private closeConnections() {
		for (const connection of this.connections_) {
			connection.abort()
		}
	}
	private showed() {
		clearTimeout(this.sleepTimer_)
		this.sleepTimer_ = null
		if (this.state_ == state.dead) {
			this.reload()
			return
		}
		this.delay_.reset()
		if (this.state_ != state.sleep) {
			return
		}
		this.state_ = state.active
		if (this.connections_.size != 0) {
			return
		}
		this.roll()
	}
	private reloaded_ = false
	private reload() {
		if (this.reloaded_) {
			return
		}
		this.reloaded_ = true
		location.reload()
	}
	suspend() {
		if (this.state_ == state.dead) {
			return
		}
		this.state_ = state.dead
		this.closeConnections();
		["pointerdown", "pointermove", "pointerup", "scroll", "focus", "keydown", "input"].forEach(event => {
			window.addEventListener(event, () => this.reload(), true);
		});
	}
	kill() {
		this.state_ = state.dead
		if (!document.hidden) {
			this.reload()
		}
		return
	}
}

const ctrl = new Controller()

export default {
	ready: ctrl.ready,
	gone() {
		ctrl.kill()
	}
}
