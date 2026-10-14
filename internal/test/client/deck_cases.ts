import { register, eq } from "./kit"
import { Package } from "../../client/package"
import { Solitaire } from "../../client/deck"

function card(spec: string): Package {
	const [label, range] = spec.split(":")
	const [beg, end] = range.split("-").map(Number)
	const seq = end == undefined ? beg : [beg, end]
	if (label == "f") {
		return new Package([seq, null])
	}
	return new Package([seq, [0], 0, ["call", label], null])
}

function put(deck: Solitaire, ...specs: Array<string>) {
	for (const spec of specs) {
		deck.insert(card(spec))
	}
}

function deal(...specs: Array<string>): Solitaire {
	const deck = new Solitaire()
	put(deck, ...specs)
	return deck
}

function ran(deck: Solitaire): Array<string> {
	return deck.collect().map(p => `${p.arg}:${p.end}`)
}

register("deck: packages arriving in order are each delivered by the next collect and never reported lost", () => {
	const deck = new Solitaire()
	put(deck, "a:1")
	eq(ran(deck), ["a:1"])
	eq(ran(deck), [])
	put(deck, "b:2", "c:3")
	eq(deck.collectLost(), [])
	eq(ran(deck), ["b:2", "c:3"])
	put(deck, "d:4")
	eq(ran(deck), ["d:4"])
	eq(deck.collectLost(), [])
})

register("deck: a package ahead of the cursor waits, its gap is reported lost once, and both collect in order when the gap closes", () => {
	const deck = new Solitaire()
	deck.insert(new Package([2, [0], 0, ["call", []], null]))
	eq(deck.collect().map(p => p.end), [])
	eq(deck.collectLost(), [1])
	eq(deck.collectLost(), [])
	deck.insert(new Package([1, [0], 0, ["call", []], null]))
	eq(deck.collect().map(p => p.end), [1, 2])
	eq(deck.collectLost(), [])
})

register("deck: collect returns nothing while the head seq is missing, however many later packages are held", () => {
	const deck = deal("c:3", "d:4", "b:2", "e:5")
	eq(ran(deck), [])
	eq(ran(deck), [])
	put(deck, "a:1")
	eq(ran(deck), ["a:1", "b:2", "c:3", "d:4", "e:5"])
})

register("deck: collect delivers the contiguous prefix only and resumes from the gap once it closes", () => {
	const deck = deal("a:1", "b:2", "d:4", "x:6")
	eq(ran(deck), ["a:1", "b:2"])
	eq(ran(deck), [])
	put(deck, "c:3")
	eq(ran(deck), ["c:3", "d:4"])
	put(deck, "e:5")
	eq(ran(deck), ["e:5", "x:6"])
	eq(ran(deck), [])
})

register("deck: packages inserted in reverse order are delivered in ascending seq order", () => {
	const deck = deal("e:5", "d:4", "c:3", "b:2", "a:1")
	eq(deck.collectLost(), [])
	eq(ran(deck), ["a:1", "b:2", "c:3", "d:4", "e:5"])
})

register("deck: an empty deck collects nothing and reports nothing lost", () => {
	const deck = new Solitaire()
	eq(ran(deck), [])
	eq(deck.collectLost(), [])
	deck.return([1, 2])
	eq(deck.collectLost(), [])
})

register("deck: lost lists every missing seq from the cursor up to the last held card, ascending and without repeats", () => {
	const deck = deal("a:1", "d:4", "h:8", "e:5")
	eq(deck.collectLost(), [2, 3, 6, 7])
	eq(ran(deck), ["a:1"])
	eq(deck.collectLost(), [])
	put(deck, "k:11")
	eq(deck.collectLost(), [9, 10])
})

register("deck: seqs after the last held card are not lost until a later card such as a probe filler arrives", () => {
	const deck = deal("a:1")
	eq(ran(deck), ["a:1"])
	eq(deck.collectLost(), [])
	put(deck, "f:4")
	eq(deck.collectLost(), [2, 3])
	eq(ran(deck), [])
})

register("deck: a gap already reported is reported again only after it was handed back by return", () => {
	const deck = deal("d:4")
	eq(deck.collectLost(), [1, 2, 3])
	eq(deck.collectLost(), [])
	put(deck, "x:6")
	eq(deck.collectLost(), [5])
	deck.return([2])
	eq(deck.collectLost(), [2])
	eq(deck.collectLost(), [])
	deck.return([1, 2, 3, 5])
	eq(deck.collectLost(), [1, 2, 3, 5])
	eq(deck.collectLost(), [])
})

register("deck: a returned seq whose gap closed meanwhile is not reported again", () => {
	const deck = deal("d:4")
	eq(deck.collectLost(), [1, 2, 3])
	put(deck, "b:2")
	deck.return([1, 2, 3])
	eq(deck.collectLost(), [1, 3])
	put(deck, "a:1", "f:3")
	deck.return([1, 3])
	eq(deck.collectLost(), [])
	eq(ran(deck), ["a:1", "b:2", "d:4"])
	deck.return([1, 2, 3])
	eq(deck.collectLost(), [])
})

register("deck: returning seqs that were never reported changes nothing", () => {
	const deck = deal("c:3")
	deck.return([1, 2, 7])
	eq(deck.collectLost(), [1, 2])
	eq(deck.collectLost(), [])
})

register("deck: a filler over a gap closes it and is not delivered as a package", () => {
	const deck = deal("a:1", "d:4")
	eq(ran(deck), ["a:1"])
	eq(deck.collectLost(), [2, 3])
	put(deck, "f:2-3")
	eq(ran(deck), ["d:4"])
	deck.return([2, 3])
	eq(deck.collectLost(), [])
})

register("deck: single seq fillers close a gap one seq at a time", () => {
	const deck = deal("d:4")
	put(deck, "f:2")
	eq(deck.collectLost(), [1, 3])
	put(deck, "f:1")
	eq(ran(deck), [])
	put(deck, "f:3")
	eq(ran(deck), ["d:4"])
})

register("deck: the cursor advances past collected fillers, so a later package is delivered and the filled seqs become stale", () => {
	const deck = deal("f:1-3")
	eq(ran(deck), [])
	eq(deck.collectLost(), [])
	put(deck, "b:2", "c:3")
	eq(ran(deck), [])
	put(deck, "d:4")
	eq(ran(deck), ["d:4"])
	put(deck, "f:5", "f:6-7", "h:8")
	eq(ran(deck), ["h:8"])
	eq(deck.collectLost(), [])
})

register("deck: a trailing filler is consumed with the prefix and the next seq follows it", () => {
	const deck = deal("a:1", "f:2-4")
	eq(ran(deck), ["a:1"])
	put(deck, "d:4")
	eq(ran(deck), [])
	put(deck, "e:5")
	eq(ran(deck), ["e:5"])
})

register("deck: a ranged card covers the seqs before its end, is delivered once as its end seq and leaves no gap", () => {
	const deck = deal("c:1-3")
	eq(deck.collectLost(), [])
	eq(ran(deck), ["c:3"])
	put(deck, "g:5-7")
	eq(deck.collectLost(), [4])
	put(deck, "d:4")
	eq(ran(deck), ["d:4", "g:7"])
})

register("deck: a ranged card arriving over a reported gap closes exactly the seqs it covers", () => {
	const deck = deal("a:1", "h:8")
	eq(deck.collectLost(), [2, 3, 4, 5, 6, 7])
	put(deck, "e:3-5")
	deck.return([2, 3, 4, 5, 6, 7])
	eq(deck.collectLost(), [2, 6, 7])
	eq(ran(deck), ["a:1"])
	put(deck, "b:2", "g:6-7")
	eq(ran(deck), ["b:2", "e:5", "g:7", "h:8"])
})

register("deck: a ranged card replaces held cards whose seqs it covers, they are not delivered", () => {
	const deck = deal("b:2", "c:3", "e:5")
	put(deck, "d:1-4")
	eq(deck.collectLost(), [])
	eq(ran(deck), ["d:4", "e:5"])
})

register("deck: a card overlapping the start of a held ranged card takes the shared seqs, both are delivered and no seq is lost", () => {
	const deck = deal("m:3-6")
	put(deck, "d:2-4")
	eq(deck.collectLost(), [1])
	put(deck, "a:1")
	eq(ran(deck), ["a:1", "d:4", "m:6"])
})

register("deck: a card overlapping the end of a held ranged card and reaching past it replaces it and keeps its start", () => {
	const deck = deal("m:3-6")
	put(deck, "q:5-8")
	eq(deck.collectLost(), [1, 2])
	put(deck, "f:1-2")
	eq(ran(deck), ["q:8"])
})

register("deck: a package landing inside a held ranged card is kept and the ranged card still delivers its end", () => {
	const deck = deal("m:2-6")
	put(deck, "d:4")
	eq(deck.collectLost(), [1])
	put(deck, "a:1")
	eq(ran(deck), ["a:1", "d:4", "m:6"])
})

register("deck: a resend that extends the start of a held card closes the gap before it and is delivered once", () => {
	const deck = deal("a:1", "d:4")
	eq(ran(deck), ["a:1"])
	eq(deck.collectLost(), [2, 3])
	put(deck, "r:2-4")
	eq(ran(deck), ["r:4"])
	eq(ran(deck), [])
	deck.return([2, 3])
	eq(deck.collectLost(), [])
})

register("deck: a resend that extends the start of a held card only part of the way leaves the rest of the gap lost", () => {
	const deck = deal("e:5")
	eq(deck.collectLost(), [1, 2, 3, 4])
	put(deck, "r:3-5")
	deck.return([1, 2, 3, 4])
	eq(deck.collectLost(), [1, 2])
	put(deck, "f:1-2")
	eq(ran(deck), ["r:5"])
})

register("deck: one wide card covers a chain of held cards, drops those it fully covers and trims the one it reaches into", () => {
	const deck = deal("c:3", "e:5", "g:7", "j:8-10", "l:12")
	put(deck, "i:2-9")
	eq(deck.collectLost(), [1, 11])
	put(deck, "a:1", "k:11")
	eq(ran(deck), ["a:1", "i:9", "j:10", "k:11", "l:12"])
})

register("deck: a filler covering a chain of held cards up to the last one empties the chain", () => {
	const deck = deal("b:2", "d:4", "x:6")
	put(deck, "f:1-6")
	eq(deck.collectLost(), [])
	eq(ran(deck), [])
	put(deck, "g:7")
	eq(ran(deck), ["g:7"])
})

register("deck: a duplicate of a held package is delivered once", () => {
	const deck = deal("b:2", "b:2", "c:3", "b:2", "c:3")
	eq(deck.collectLost(), [1])
	put(deck, "a:1", "a:1")
	eq(ran(deck), ["a:1", "b:2", "c:3"])
	eq(ran(deck), [])
})

register("deck: stale packages and fillers below the cursor are ignored, not delivered again and never reported lost", () => {
	const deck = deal("a:1", "b:2", "c:3")
	eq(ran(deck), ["a:1", "b:2", "c:3"])
	put(deck, "a:1", "c:3", "x:1-2", "f:2", "f:1-3")
	eq(ran(deck), [])
	eq(deck.collectLost(), [])
	put(deck, "d:4")
	eq(ran(deck), ["d:4"])
	eq(deck.collectLost(), [])
})

register("deck: a ranged card that starts below the cursor and ends at or after it is delivered once", () => {
	const deck = deal("a:1", "b:2")
	eq(ran(deck), ["a:1", "b:2"])
	put(deck, "d:1-4")
	eq(deck.collectLost(), [])
	eq(ran(deck), ["d:4"])
	put(deck, "f:2-6", "g:7")
	eq(ran(deck), ["g:7"])
	eq(deck.collectLost(), [])
})

register("deck: a ranged card starting below the cursor while later cards are held joins them without a gap", () => {
	const deck = deal("a:1", "b:2", "x:6")
	eq(ran(deck), ["a:1", "b:2"])
	eq(deck.collectLost(), [3, 4, 5])
	put(deck, "e:2-5")
	eq(ran(deck), ["e:5", "x:6"])
})

register("deck: after everything held was collected the deck accepts and delivers further packages", () => {
	const deck = deal("a:1")
	eq(ran(deck), ["a:1"])
	eq(ran(deck), [])
	put(deck, "c:3")
	eq(ran(deck), [])
	eq(deck.collectLost(), [2])
	put(deck, "b:2")
	eq(ran(deck), ["b:2", "c:3"])
	put(deck, "d:4")
	eq(ran(deck), ["d:4"])
})

type Row = [name: string, held: Array<string>, insert: string, lost: Array<number>, delivered: Array<string>]

const rows: Array<Row> = [
	["before the only card", ["a:5"], "p:3", [1, 2, 4], ["p:3", "a:5"]],
	["before the first of a chain", ["a:5", "b:7"], "p:3", [1, 2, 4, 6], ["p:3", "a:5", "b:7"]],
	["ranged before the first, touching it", ["a:5", "b:7"], "p:2-4", [1, 6], ["p:4", "a:5", "b:7"]],
	["after the only card", ["a:3"], "p:5", [1, 2, 4], ["a:3", "p:5"]],
	["between two cards", ["a:3", "b:7"], "p:5", [1, 2, 4, 6], ["a:3", "p:5", "b:7"]],
	["after the last of a chain", ["a:3", "b:5"], "p:7", [1, 2, 4, 6], ["a:3", "b:5", "p:7"]],
	["same seq replaces the held card", ["a:3"], "p:3", [1, 2], ["p:3"]],
	["same seq keeps the start of a held range", ["a:2-3"], "p:3", [1], ["p:3"]],
	["start on a held end takes over its range", ["a:2-3"], "p:3-5", [1], ["p:5"]],
	["start on a held end, next fully covered is dropped", ["a:3", "b:5"], "p:3-6", [1, 2], ["p:6"]],
	["start on a held end, next partly covered is trimmed", ["a:3", "b:5-8"], "p:3-6", [1, 2], ["p:6", "b:8"]],
	["start on a held end, next beyond the new end is untouched", ["a:3", "b:8"], "p:3-6", [1, 2, 7], ["p:6", "b:8"]],
	["start on a held end, next touching the new end is untouched", ["a:3", "b:7"], "p:3-6", [1, 2], ["p:6", "b:7"]],
	["identical range replaces the held card", ["a:3-5"], "p:3-5", [1, 2], ["p:5"]],
	["wider on both sides replaces the held card", ["a:4-5"], "p:3-7", [1, 2], ["p:7"]],
	["wider at the start, same end", ["a:4-5"], "p:2-5", [1], ["p:5"]],
	["same start, longer end", ["a:3-5"], "p:3-7", [1, 2], ["p:7"]],
	["wider card drops covered cards and trims the reached one", ["a:3", "b:5", "c:7", "d:8-10"], "p:2-9", [1], ["p:9", "d:10"]],
	["wider card drops every later card it covers", ["a:3", "b:5", "c:7"], "p:2-7", [1], ["p:7"]],
	["wider card stops at a later card beyond its end", ["a:3", "b:5", "c:9"], "p:2-6", [1, 7, 8], ["p:6", "c:9"]],
	["start inside, end beyond keeps the held start", ["a:3-5"], "p:4-7", [1, 2], ["p:7"]],
	["start inside, same end keeps the held start", ["a:3-5"], "p:4-5", [1, 2], ["p:5"]],
	["start inside, end on the next card drops it", ["a:3-5", "b:7"], "p:4-7", [1, 2], ["p:7"]],
	["start inside, end inside the next card trims it", ["a:3-5", "b:7-9"], "p:4-8", [1, 2], ["p:8", "b:9"]],
	["inside a held range splits it", ["a:3-6"], "p:4", [1, 2], ["p:4", "a:6"]],
	["ranged inside a held range splits it", ["a:3-7"], "p:4-5", [1, 2], ["p:5", "a:7"]],
	["on the start of a held range splits it", ["a:3-6"], "p:3", [1, 2], ["p:3", "a:6"]],
	["start before, end inside a held range", ["a:4-6"], "p:2-4", [1], ["p:4", "a:6"]],
	["inside a held range with a card after it", ["a:3-6", "b:8"], "p:4", [1, 2, 7], ["p:4", "a:6", "b:8"]],
	["inside a held filler", ["f:3-6"], "p:5", [1, 2], ["p:5"]],
	["inside the second card, bridging the gap before it", ["a:1", "b:5-8"], "p:3-6", [2], ["a:1", "p:6", "b:8"]],
	["from inside the first card to inside the second", ["a:2-3", "b:6-8"], "p:3-7", [1], ["p:7", "b:8"]],
	["filler over a held package voids it", ["a:4"], "f:3-6", [1, 2], []],
	["filler on the seq of a held package voids it", ["a:3"], "f:3", [1, 2], []],
	["package on the seq of a held filler is delivered", ["f:3"], "p:3", [1, 2], ["p:3"]],
	["filler inside a held range keeps the held end", ["a:3-6"], "f:4-5", [1, 2], ["a:6"]],
	["filler before, touching a held card", ["a:4"], "f:1-3", [], ["a:4"]],
	["filler after, touching a held card", ["a:1"], "f:2-3", [], ["a:1"]],
]

for (const [name, held, insert, lost, delivered] of rows) {
	register(`deck insert: ${name}`, () => {
		const deck = deal(...held)
		put(deck, insert)
		const got = deck.collectLost()
		eq(got, lost)
		eq(deck.collectLost(), [])
		for (const seq of got) {
			put(deck, `f:${seq}`)
		}
		eq(ran(deck), delivered)
		eq(ran(deck), [])
		deck.return(got)
		eq(deck.collectLost(), [])
	})
}

class Model {
	private state = new Map<number, string | null>()
	private reported = new Set<number>()
	private cursor = 1
	insert(beg: number, end: number, label: string | null) {
		if (end < this.cursor) {
			return
		}
		for (let seq = beg; seq < end; seq++) {
			this.state.set(seq, null)
		}
		this.state.set(end, label)
	}
	collect(): Array<string> {
		const a: Array<string> = []
		while (this.state.has(this.cursor)) {
			const label = this.state.get(this.cursor)
			if (label != null) {
				a.push(`${label}:${this.cursor}`)
			}
			this.state.delete(this.cursor)
			this.cursor += 1
		}
		return a
	}
	collectLost(): Array<number> {
		const lost: Array<number> = []
		const last = Math.max(0, ...this.state.keys())
		for (let seq = this.cursor; seq < last; seq++) {
			if (this.state.has(seq) || this.reported.has(seq)) {
				continue
			}
			lost.push(seq)
			this.reported.add(seq)
		}
		return lost
	}
	return(lost: Array<number>) {
		for (const seq of lost) {
			this.reported.delete(seq)
		}
	}
}

function random(seed: number): (n: number) => number {
	let s = seed
	return n => {
		s = (Math.imul(s, 1664525) + 1013904223) >>> 0
		return Math.floor(s / 4294967296 * n)
	}
}

register("deck: random inserts, collects, lost reports and returns match a per seq model where the latest card wins", () => {
	for (let seed = 1; seed <= 3000; seed++) {
		const rnd = random(seed)
		const deck = new Solitaire()
		const model = new Model()
		const log: Array<string> = []
		let out: Array<number> = []
		try {
			for (let step = 0; step < 40; step++) {
				const op = rnd(10)
				if (op < 6) {
					const end = 1 + rnd(14)
					const beg = Math.max(1, end - (rnd(3) == 0 ? rnd(5) : 0))
					const label = rnd(4) == 0 ? "f" : `p${step}`
					const spec = `${label}:${beg}-${end}`
					log.push(spec)
					deck.insert(card(spec))
					model.insert(beg, end, label == "f" ? null : label)
					continue
				}
				if (op < 8) {
					log.push("collect")
					eq(ran(deck), model.collect())
					continue
				}
				if (op < 9) {
					log.push("lost")
					const lost = deck.collectLost()
					eq(lost, model.collectLost())
					out.push(...lost)
					continue
				}
				const back = out.filter(() => rnd(2) == 0)
				out = out.filter(seq => !back.includes(seq))
				log.push(`return ${back.join(",")}`)
				deck.return(back)
				model.return(back)
			}
			log.push("drain")
			eq(deck.collectLost(), model.collectLost())
			eq(ran(deck), model.collect())
		} catch (e: any) {
			throw new Error(`seed ${seed}: ${log.join(" | ")}: ${e.message}`)
		}
	}
})
