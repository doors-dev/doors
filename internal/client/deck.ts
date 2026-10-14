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
import type { Lost } from "./connector";
import { STRESS_MODE } from "./lib";


export class Solitaire {
	top: Card
	private cursor: number = 1
	constructor() {

	}
	private collectedLost = new Set<number>()
	return(lost: Lost) {
		for (const gap of lost) {
			this.collectedLost.delete(gap)
		}
	}
	collectLost(): Lost {
		const lost: Lost = []
		for (const seq of this.getLost()) {
			if (this.collectedLost.has(seq)) {
				continue
			}
			lost.push(seq)
			this.collectedLost.add(seq)
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
		this.top.lost(this.cursor - 1, lost)
		return lost
	}
	collect(): Array<Package> {
		const a: Array<Package> = []
		if (!this.top) {
			return a
		}
		if (this.top.start > this.cursor) {
			return a
		}
		this.top.collect(a, this)
		const tail = a[a.length - 1]
		this.cursor = tail.end + 1
		return a.filter(p => !p.isFiller)
	}
	insert(p: Package) {
		if (p.end < this.cursor) {
			return
		}
		if (!this.top) {
			this.top = new Card(p)
			return
		}
		this.top.insert(p)
	}

}

export class Card {
	private next: Card
	constructor(private p: Package) {
	}
	get end(): number {
		return this.p.end
	}
	get start(): number {
		return this.p.start
	}
	lost(end: number, lost: Lost) {
		if (end < this.p.start - 1) {
			const lostStart = end + 1
			const lostEnd = this.p.start - 1
			for (let seq = lostStart; seq <= lostEnd; seq++) {
				lost.push(seq)
			}
		}
		if (this.next) {
			this.next.lost(this.p.end, lost)
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
		a.push(this.p)
		h.top = this.next
		if (this.next) {
			this.next.collect(a, h)
		}
	}
	insert(p: Package) {
		if (p.end < this.p.start) {
			if (this.next) {
				this.next.insert(this.p)
			} else {
				this.next = new Card(this.p)
			}
			this.p = p
			return
		}
		if (p.start > this.p.end) {
			if (this.next) {
				this.next.insert(p)
			} else {
				this.next = new Card(p)
			}
			return
		}
		if (p.start == this.p.end) {
			p.start = this.p.start
			this.p = p
			if (this.next) {
				this.next.cover(this.p.end, this)
			}
			return
		}
		// start < end
		if (p.start <= this.p.start && p.end >= this.p.end) {
			this.p = p
			if (this.next) {
				this.next.cover(this.p.end, this)
			}
			return
		}
		if (p.end >= this.p.end) { // && p.start > this.start
			p.start = this.p.start
			this.p = p
			if (this.next) {
				this.next.cover(this.p.end, this)
			}
			return
		}
		// p.end < this.end
		p.start = Math.min(this.p.start, p.start)
		this.p.start = p.end + 1
		if (this.next) {
			this.next.insert(this.p)
		} else {
			this.next = new Card(this.p)
		}
		this.p = p
	}
	private cover(end: number, head: Card) {
		if (end >= this.p.end) {
			head.next = this.next
			if (this.next) {
				this.next.cover(end, head)
			}
			return
		}
		if (this.p.start > end) {
			return
		}
		this.p.start = end + 1
		return
	}
}
