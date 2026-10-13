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

import { Package } from "./package";
import type { Results } from "./connector";
import { Result } from "./lib";

type Outcome = Result<any, { message: string }>

export type Run = (p: Package) => Outcome | Promise<Outcome>


export class Batch {
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

export class Tracker {
	private buffered_: Results = new Map()
	constructor(private notify_: () => void, private run_: Run, private pending_: number) {
	}
	private batches = new Set<Batch>()
	private idLookup = new Map<number, Batch>()
	private doorLookup = new Map<number, Batch>()
	private packageLookup = new Map<number, Batch>()
	private tagLookup = new Map<number, Set<Batch>>()
	removeBatch(batch: Batch) {
		this.batches.delete(batch)
		if (this.idLookup.get(batch.id) == batch) {
			this.idLookup.delete(batch.id)
		}
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
		this.packageLookup.set(p.end, batch)
		if (p.door != 0) {
			this.doorLookup.set(p.door, batch)
		}
	}
	unhold(p: Package, batch: Batch) {
		this.packageLookup.delete(p.end)
		if (this.doorLookup.get(p.door) == batch) {
			this.doorLookup.delete(p.door)
		}
	}
	push(p: Package) {
		this.process(p)
		while (this.packageLookup.size >= this.pending_) {
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
		const result = this.run_(p)
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
