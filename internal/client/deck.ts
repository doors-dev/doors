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

export class Card {
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
