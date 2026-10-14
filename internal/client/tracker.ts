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
	constructor(private tracker: Tracker, private finished: boolean, readonly id = 0, readonly group = 0) {
	}
	empty(): boolean {
		return this.packages.length == 0
	}
	finish(closer: Package) {
		this.finished = true
		this.packages.push(closer)
		this.tracker.register(closer, this)
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
	wait(batch: Batch | undefined) {
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
	private groupLookup = new Map<number, Set<Batch>>()
	private reportLookup = new Map<number, Batch>()
	private held = 0
	removeBatch(batch: Batch) {
		this.batches.delete(batch)
		if (this.idLookup.get(batch.id) == batch) {
			this.idLookup.delete(batch.id)
		}
		this.untag(batch.group, batch)
		if (this.reportLookup.get(batch.group) == batch) {
			this.reportLookup.delete(batch.group)
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
		this.register(p, batch)
	}
	register(p: Package, batch: Batch) {
		this.held += 1
		this.tag(p.group, batch)
	}
	unhold(p: Package, batch: Batch) {
		this.held -= 1
		this.packageLookup.delete(p.end)
		if (this.doorLookup.get(p.door) == batch) {
			this.doorLookup.delete(p.door)
		}
		this.untag(p.group, batch)
	}
	private tag(group: number, batch: Batch) {
		if (group == 0) {
			return
		}
		let tagged = this.groupLookup.get(group)
		if (tagged == undefined) {
			tagged = new Set()
			this.groupLookup.set(group, tagged)
		}
		tagged.add(batch)
	}
	private untag(group: number, batch: Batch) {
		const tagged = this.groupLookup.get(group)
		if (tagged == undefined) {
			return
		}
		tagged.delete(batch)
		if (tagged.size == 0) {
			this.groupLookup.delete(group)
		}
	}
	private oldest(): (Batch | undefined) {
		for (const batch of this.batches) {
			if (!batch.empty()) {
				return batch
			}
		}
		return undefined
	}
	push(p: Package) {
		this.process(p)
		while (this.held >= this.pending_) {
			const batch = this.oldest()
			if (batch == undefined) {
				break
			}
			this.flush(batch, true)
		}
	}
	private process(p: Package) {
		if (p.action == "batch_init") {
			if (this.idLookup.has(p.batch)) {
				this.buffered_.set(p.end, [undefined, "batch is already intialized"])
				return
			}
			const batch = new Batch(this, false, p.batch, p.group)
			this.batches.add(batch)
			this.idLookup.set(p.batch, batch)
			this.tag(p.group, batch)
			this.reportLookup.get(p.group)?.wait(batch)
			this.execute(p)
			return
		}
		const report = p.action == "report_hook"
		const closer = p.batch > 0 && (p.action == "batch_flush" || report)
		let batch = p.batch > 0 ? this.idLookup.get(p.batch) : undefined
		switch (true) {
			case batch != undefined && !closer:
				batch.add(p)
				return
			case batch != undefined:
				this.idLookup.delete(p.batch)
				batch.finish(p)
				break
			default:
				batch = new Batch(this, true, 0, p.group)
				this.batches.add(batch)
				batch.add(p)
		}
		if (report) {
			batch.waitAll(this.groupLookup.get(p.group))
			this.reportLookup.set(p.group, batch)
		}
		this.flush(batch)
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
