import { Package } from "../../client/package"
import { Tracker } from "../../client/tracker"

type Case = () => void | Promise<void>

type Link = { b?: number, s?: number, d?: number }

export const cases = new Map<string, Case>()

export function register(name: string, fn: Case) {
	if (cases.has(name)) {
		throw new Error(`duplicate case: ${name}`)
	}
	cases.set(name, fn)
}

export function eq(got: any, want: any) {
	const g = JSON.stringify(got)
	const w = JSON.stringify(want)
	if (g != w) {
		throw new Error(`got ${g}, want ${w}`)
	}
}

export class Stream {
	ran: Array<number> = []
	private seq = 0
	private tracker: Tracker
	constructor(pending = 1000) {
		this.tracker = new Tracker(() => { }, p => {
			this.ran.push(p.end)
			return [null, undefined]
		}, pending)
	}
	push(action: string, door: number, link: Link = {}): Array<number> {
		const from = this.ran.length
		this.seq += 1
		this.tracker.push(new Package([this.seq, [link.b ?? 0, door, link.s ?? 0, link.d ?? 0], [action, []], null]))
		return this.ran.slice(from)
	}
	init(batch: number): Array<number> {
		return this.push("batch_init", 0, { b: batch })
	}
	call(door: number, link: Link = {}): Array<number> {
		return this.push("call", door, link)
	}
	flush(batch: number): Array<number> {
		return this.push("batch_flush", 0, { b: batch })
	}
	hook(batch: number, door: number): Array<number> {
		return this.push("report_hook", door, { b: batch })
	}
	errors(): Array<number> {
		return [...this.tracker.collect()].filter(([, r]) => r[1] !== undefined).map(([seq]) => seq)
	}
}

export function leftovers(s: Stream): Array<string> {
	const tracker: any = s["tracker"]
	return ["batches", "idLookup", "doorLookup", "packageLookup", "tagLookup"].filter(k => tracker[k].size != 0)
}

export async function fails(fn: () => any, want?: string) {
	let got: string | undefined
	try {
		await fn()
	} catch (e: any) {
		got = String(e?.message ?? e)
	}
	if (got === undefined) {
		throw new Error(`got no error, want ${want ?? "an error"}`)
	}
	if (want !== undefined) {
		eq(got, want)
	}
}
