import { register, eq, leftovers, Stream } from "./kit"

type Link = { b?: number, s?: number, d?: number }

type Op = (s: Stream) => Array<number>

type Row = { name: string, cap?: number, errors?: Array<number>, steps: Array<[Op, Array<number>]> }

const init = (b: number): Op => s => s.init(b)
const c = (t: number, link: Link = {}): Op => s => s.call(t, link)
const fl = (b: number, t = 0): Op => s => s.push("batch_flush", t, { b })
const hook = (b: number, t: number): Op => s => s.hook(b, t)

const rows: Array<Row> = [
	{ name: "plain traffic runs at once", steps: [[c(1), [1]], [c(2), [2]], [c(1), [3]], [c(1, { s: 1, d: 2 }), [4]]] },
	{ name: "hold and release", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2, { b: 1 }), []], [fl(1), [2, 3, 4]]] },
	{ name: "independent door overtakes a held batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2), [3]], [fl(1), [2, 4]]] },
	{ name: "same-door packages wait behind the batch holding the door", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1), []], [c(1), []], [c(2), [5]], [fl(1), [2, 3, 4, 6]]] },
	{ name: "chaining through a waiting holder", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2, { s: 2 }), []], [c(2), []], [c(3), [5]], [c(3, { d: 2 }), []], [fl(1), [2, 3, 4, 6, 7]]] },
	{ name: "S wait behind a held package", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(5, { s: 2 }), []], [c(6), [4]], [fl(1), [2, 3, 5]]] },
	{ name: "S already executed, plain", steps: [[c(1), [1]], [c(2, { s: 1 }), [2]]] },
	{ name: "S already executed, was held and released", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(1), [2, 3]], [c(3, { s: 2 }), [4]]] },
	{ name: "D wait, Static shape with T parent and D replaced", steps: [[init(1), [1]], [c(7, { b: 1 }), []], [c(3, { d: 7 }), []], [c(3), []], [c(4), [5]], [fl(1), [2, 3, 4, 6]]] },
	{ name: "Static as first call with S and D, S released first", steps: [[init(1), [1]], [init(2), [2]], [c(4, { b: 1 }), []], [c(7, { b: 2 }), []], [c(5, { s: 3, d: 7 }), []], [fl(1), [3, 6]], [fl(2), [4, 5, 7]]] },
	{ name: "Static as first call with S and D, D released first", steps: [[init(1), [1]], [init(2), [2]], [c(4, { b: 1 }), []], [c(7, { b: 2 }), []], [c(5, { s: 3, d: 7 }), []], [fl(2), [4, 6]], [fl(1), [3, 5, 7]]] },
	{ name: "closer batch_flush runs last in its batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(1), [2, 3]]] },
	{ name: "closer report_hook with batch runs last in its batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [hook(1, 2), [2, 3]]] },
	{ name: "report_hook without batch is not a closer", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [hook(0, 2), [3]], [hook(0, 1), []], [fl(1), [2, 4, 5]]] },
	{ name: "batch_flush with batch 0 is a plain package", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(0, 1), []], [fl(0, 2), [4]], [fl(1), [2, 3, 5]]] },
	{ name: "closer blocked behind another batch", steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [fl(2), []], [c(9), [6]], [fl(1), [3, 4, 5, 7]]] },
	{ name: "transitive chain of three batches released at once", steps: [[init(1), [1]], [init(2), [2]], [init(3), [3]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [c(2, { b: 2 }), []], [c(2, { b: 3 }), []], [fl(3), []], [fl(2), []], [fl(1), [4, 5, 6, 7, 8, 9, 10]]] },
	{ name: "cycle of two batches releases together", steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [c(2, { b: 2 }), []], [c(2, { b: 1 }), []], [fl(1), []], [fl(2), [3, 4, 5, 6, 7, 8]]] },
	{ name: "cycle of two batches with a waiter on the cycle", steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [c(2, { b: 2 }), []], [c(2, { b: 1 }), []], [c(2), []], [fl(2), []], [fl(1), [3, 4, 5, 6, 7, 8, 9]]] },
	{ name: "cycle of three batches releases together", steps: [[init(1), [1]], [init(2), [2]], [init(3), [3]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [c(2, { b: 2 }), []], [c(2, { b: 3 }), []], [c(3, { b: 3 }), []], [c(3, { b: 1 }), []], [fl(1), []], [fl(2), []], [fl(3), [4, 5, 6, 7, 8, 9, 10, 11, 12]]] },
	{ name: "members of the same batch depend on each other", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 1 }), []], [c(2, { b: 1, s: 2 }), []], [c(3, { b: 1, d: 1 }), []], [fl(1), [2, 3, 4, 5, 6]]] },
	{ name: "duplicate batch_init answered with an error, batch intact", errors: [2], steps: [[init(1), [1]], [init(1), []], [c(1, { b: 1 }), []], [fl(1), [3, 4]]] },
	{ name: "id reuse after the old batch fully released", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(1), [2, 3]], [init(1), [4]], [c(1, { b: 1 }), []], [c(2), [6]], [fl(1), [5, 7]]] },
	{ name: "id reuse while the old batch is closed but still blocked", steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 2 }), []], [c(1, { b: 1 }), []], [fl(1), []], [init(1), [6]], [c(5, { b: 1 }), []], [fl(2), [3, 4, 5, 8]], [c(6, { b: 1 }), []], [fl(1), [7, 9, 10]]] },
	{ name: "soft never-initialised tag runs at once", steps: [[c(1, { b: 5 }), [1]], [c(1, { b: 5 }), [2]]] },
	{ name: "soft closer waits for a held member of its tag", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 5 }), []], [fl(5), []], [c(2, { b: 6 }), [5]], [fl(1), [2, 3, 4, 6]]] },
	{ name: "soft closer waits only for its own tag", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 5 }), []], [fl(6), [4]], [hook(5, 8), []], [fl(1), [2, 3, 5, 6]]] },
	{ name: "soft free member does not delay the closer", steps: [[c(1, { b: 5 }), [1]], [fl(5), [2]]] },
	{ name: "soft closer with nothing queued runs at once", steps: [[hook(7, 3), [1]]] },
	{ name: "soft member arriving after its closer", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 5 }), []], [fl(5), []], [c(2, { b: 5 }), [5]], [c(1, { b: 5 }), []], [fl(1), [2, 3, 4, 6, 7]]] },
	{ name: "member arriving after its hard batch closed runs at once", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(1), [2, 3]], [c(1, { b: 1 }), [4]]] },
	{ name: "soft closer waits behind the batch holding its own door", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(5, 1), []], [fl(1), [2, 3, 4]]] },
	{ name: "cap reached by members, late member and late closer run at once", cap: 2, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2, { b: 1 }), [2, 3]], [c(1, { b: 1 }), [4]], [fl(1), [5]]] },
	{ name: "cap skips an older empty batch, which stays open and holds its later members", cap: 2, steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 2 }), []], [c(2, { b: 2 }), [3, 4]], [c(3, { b: 1 }), []], [fl(1), [5, 6]], [fl(2), [7]]] },
	{ name: "cap forces a batch together with an open parent created later", cap: 3, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2), [3]], [c(2, { b: 2 }), []], [c(2, { b: 1 }), [2, 4, 5]], [c(3, { b: 2 }), [6]], [fl(2), [7]], [fl(1), [8]]] },
	{ name: "cap forced batch, held late member and late closer waiting for it", cap: 4, steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(5, { b: 2 }), []], [c(2, { b: 1 }), []], [c(3, { b: 1 }), [3, 5, 6]], [c(5, { b: 1 }), []], [fl(1), []], [c(9), [9]], [fl(2), [4, 7, 8, 10]]] },
	{ name: "cap reached by waiters, not members", cap: 3, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1), []], [c(1), [2, 3, 4]], [fl(1), [5]]] },
	{ name: "cap forced batch does not release its child early", cap: 3, steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(1, { b: 2 }), []], [c(3, { b: 1 }), [3, 5]], [c(1), []], [fl(2), [4, 6, 7]], [fl(1), [8]]] },
	{ name: "cap of 1 disables holding", cap: 1, steps: [[c(1), [1]], [init(1), [2]], [c(1, { b: 1 }), [3]], [c(1, { b: 1 }), [4]], [fl(1), [5]]] },
	{ name: "door 0 plain packages run at once", steps: [[c(0), [1]], [c(0), [2]]] },
	{ name: "held door 0 member does not block packages without a dependency door", steps: [[init(1), [1]], [c(0, { b: 1 }), []], [c(1), [3]], [c(2, { b: 7 }), [4]], [fl(1), [2, 5]]] },
	{ name: "held package's D is another package's T, no wait behind the D owner", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { d: 2 }), []], [c(2), [4]], [fl(1), [2, 3, 5]]] },
	{ name: "held T and held D on one door, a third package waits for the T holder only", steps: [[init(1), [1]], [init(2), [2]], [c(2, { b: 1 }), []], [c(5, { b: 2 }), []], [c(5, { d: 2 }), []], [c(2), []], [fl(1), [3, 6, 7]], [fl(2), [4, 5, 8]]] },
]

for (const row of rows) {
	register(`tracker: ${row.name}`, () => {
		const s = new Stream(row.cap)
		eq(row.steps.map(([op]) => op(s)), row.steps.map(([, want]) => want))
		const errors = row.errors ?? []
		eq(s.errors(), errors)
		eq(s.ran.length + errors.length, row.steps.length)
		eq(leftovers(s), [])
	})
}
