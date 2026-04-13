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


export const doAfter = (f: () => void) => {
	setTimeout(f, 0)
}

export const setDiff = (a: Set<string>, b: Set<string>): [Array<string>, Array<string>] => {
	const inANotB = [...a].filter(item => !b.has(item))
	const inBNotA = [...b].filter(item => !a.has(item))
	return [inANotB, inBNotA]
}

export const arrayDiff = (a: Array<string>, b: Array<string>): [Array<string>, Array<string>] => {
	const setA = new Set(a)
	const setB = new Set(b)
	return setDiff(setA, setB)
}

export const arraysEqual = (a: Array<string>, b: Array<string>, ordered = true) => {
	if (a === b) {
		return true
	}
	if (!a || !b) {
		return false
	}
	return a.length === b.length && a.every((val, i) => ordered ? val === b[i] : b.includes(val))
}

export const date = (date: Date) => {
	const pad = (num: number, l = 2) => String(num).padStart(l, "0")
	const year = date.getFullYear()
	const month = pad(date.getMonth() + 1)
	const day = pad(date.getDate())
	const hours = pad(date.getHours())
	const minutes = pad(date.getMinutes())
	const seconds = pad(date.getSeconds())
	const milliseconds = pad(date.getMilliseconds(), 3)
	const timeString = `${year}-${month}-${day}T${hours}:${minutes}:${seconds}.${milliseconds}`
	const offset = Math.abs(date.getTimezoneOffset())
	const offsetMinutes = offset % 60
	const offsetHours = Math.floor(offset / 60)
	const offsetSign = date.getTimezoneOffset() >= 0 ? "-" : "+"
	const offsetString = offsetSign + pad(offsetHours) + ":" + pad(offsetMinutes)
	return timeString + offsetString
}

export const elementId = (doorId: number) => {
	return `d0r${doorId}`
}

export const doorId = (doorId: string) => {
	return Number(doorId.slice(3))
}

export const fetchOptForm = (data: FormData) => {
	return {
		body: data,
		headers: {},
	}
}

export const fetchOptJson = (data: any) => {
	return {
		body: JSON.stringify(data),
		headers: { "Content-Type": "application/json;charset=UTF-8" }
	}
}

export const fetchOpt = (data: any) => {
	const result = {
		body: null as any,
		headers: {} as { [key: string]: string },
	}
	if (data === undefined) {
		return result
	}
	if (data instanceof FormData) {
		result.body = data
		return result
	}
	if (data instanceof URLSearchParams) {
		result.body = data
		result.headers["Content-Type"] = "application/x-www-form-urlencoded;charset=UTF-8"
		return result
	}
	if (data instanceof Blob) {
		result.body = data
		if (data.type) {
			result.headers["Content-Type"] = data.type
		}
		return result
	}
	if (data instanceof File) {
		result.body = data
		result.headers["Content-Type"] = data.type || "application/octet-stream"
		return result
	}
	if (typeof ReadableStream !== "undefined" && data instanceof ReadableStream) {
		result.body = data
		result.headers["Content-Type"] = "application/octet-stream"
		return result
	}
	if (
		data instanceof ArrayBuffer ||
		ArrayBuffer.isView(data)
	) {
		result.body = data
		result.headers["Content-Type"] = "application/octet-stream"
		return result
	}
	result.body = JSON.stringify(data)
	result.headers["Content-Type"] = "application/json;charset=UTF-8"
	return result
}

export const scrollInto = (selector: string, options?: any): boolean => {
	if(!options && options !== false) {
		options = {}
	}
	if (selector.startsWith("#")) {
		selector = "#" + CSS.escape(selector.slice(1))
	}
	const elem = document.querySelector(selector)
	if (!elem) {
		return false
	}
	elem.scrollIntoView(options);
	return true
}

export const randDelay = (): Promise<void> => {
	const min = 50
	const max = 300
	const delay = Math.floor(Math.random() * (max - min + 1)) + min
	return new Promise(resolve => setTimeout(resolve, delay))
}


export const splitClass = (str: string | undefined): Array<string> => {
	if (!str) {
		return []
	}
	return str.split(" ").map(str => str.trim()).filter(str => !!str)
}

const delayReset = 1000
const maxDelay = 12_000
const zeroThreshold = 300
const step = 200
const jitterMult = 0.4

export class ProgressiveDelay {
	private marker_ = 0
	private fee = 0
	private limited_ = false
	private resetMarker() {
		this.marker_ = Date.now()
	}
	private resetFee() {
		this.fee = 0
		this.limited_ = false
	}
	private increaseFee() {
		if (this.limited_) {
			return
		}
		this.fee++
	}
	private diff() {
		const diff = Date.now() - this.marker_
		if (diff <= zeroThreshold) {
			return 0
		}
		return diff
	}
	private delay(): number {
		let delay = step * Math.pow(2, this.fee)
		if (delay > maxDelay) {
			this.limited_ = true
			delay = maxDelay
		}
		const jitter = Math.random() * delay * jitterMult
		return delay - delay * (jitterMult / 2) + jitter
	}
	reset() {
		this.resetFee()
		this.resetMarker()
	}
	wait(): Promise<void> {
		return new Promise(res => {
			const diff = this.diff()
			if (diff >= delayReset) {
				this.resetMarker()
				this.resetFee()
				res()
				return
			}
			if (diff == 0) {
				this.increaseFee()
			}
			setTimeout(() => {
				this.resetMarker()
				res()
			}, this.delay())
		})
	}
}

export class ReliableTimer {
	private interval_: number
	private done: boolean = false
	private deadline_: number
	private tick: number
	constructor(private timeout_: number, handler: Function) {
		this.timeout_ = timeout_
		this.tick = 0.05 * this.timeout_
		this.reset()
		this.interval_ = setInterval(() => {
			if (Date.now() < this.deadline_) {
				return
			}
			this.done = true
			clearInterval(this.interval_)
			handler()
		}, this.tick)
	}
	reset() {
		this.deadline_ = Date.now() + this.timeout_ - this.tick / 2
	}
	cancel() {
		clearInterval(this.interval_)
		return !this.done
	}
}

export class AbortTimer {
	private abortController_ = new AbortController()
	private timer_: ReliableTimer
	private _expired = false
	constructor(timeout: number) {
		this.timer_ = new ReliableTimer(timeout, () => {
			if (this.signal.aborted) {
				return
			}
			this._expired = true
			this.abortController_.abort("timeout")
		})
	}
	get status(): "running" | "aborted" | "expired" {
		if (!this.signal.aborted) {
			return "running"
		}
		if (this._expired) {
			return "expired"
		}
		return "aborted"
	}
	cancel() {
		this.timer_.cancel()
	}
	abort() {
		this.timer_.cancel()
		this.abortController_.abort()
	}
	get signal() {
		return this.abortController_.signal
	}
}

export interface ReadonlySet<T> {
	readonly size: number;
	has(value: T): boolean;
	entries(): IterableIterator<[T, T]>;
	keys(): IterableIterator<T>;
	values(): IterableIterator<T>;
	forEach(callbackfn: (value: T, value2: T, set: ReadonlySet<T>) => void, thisArg?: any): void;
	[Symbol.iterator](): IterableIterator<T>;
}

