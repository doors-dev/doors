import { register, eq, leftovers, Stream, Link } from "./kit"

type Op = (s: Stream) => Array<number>

type Row = { name: string, cap?: number, errors?: Array<number>, steps: Array<[Op, Array<number>]> }

const init = (b: number, g = 0): Op => s => s.init(b, g)
const c = (t: number, link: Link = {}): Op => s => s.call(t, link)
const fl = (b: number, t = 0, g = 0): Op => s => s.push("batch_flush", t, { g, b })
const hook = (t: number, link: Link = {}): Op => s => s.hook(t, link)

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
	{ name: "closer report_hook with batch runs last in its batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [hook(2, { b: 1 }), [2, 3]]] },
	{ name: "report_hook without batch is not a closer", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [hook(2), [3]], [hook(1), []], [fl(1), [2, 4, 5]]] },
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
	{ name: "member of a batch that was never opened runs at once", steps: [[c(1, { b: 5 }), [1]], [c(1, { b: 5 }), [2]]] },
	{ name: "group-only packages and their report run at once", steps: [[c(1, { g: 5 }), [1]], [c(1, { g: 5 }), [2]], [hook(1, { g: 5 }), [3]]] },
	{ name: "report waits for a held member of its group", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { g: 5 }), []], [hook(8, { g: 5 }), []], [c(2, { g: 6 }), [5]], [fl(1), [2, 3, 4, 6]]] },
	{ name: "closer of a batch that is not open waits for a held member carrying its batch id", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 5 }), []], [fl(5), []], [c(2, { b: 6 }), [5]], [fl(1), [2, 3, 4, 6]]] },
	{ name: "report and late closer wait only for their own ids", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { g: 5 }), []], [fl(6), [4]], [hook(8, { g: 6 }), [5]], [hook(8, { g: 5 }), []], [fl(1), [2, 3, 6, 7]]] },
	{ name: "free member does not delay the late closer", steps: [[c(1, { b: 5 }), [1]], [fl(5), [2]]] },
	{ name: "report with nothing queued runs at once", steps: [[hook(3, { g: 7 }), [1]]] },
	{ name: "report waits for a member of its group held inside a blocked batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2, 9), [3]], [c(1, { g: 9, b: 2 }), []], [fl(2, 0, 9), []], [hook(6, { g: 9 }), []], [fl(1), [2, 4, 5, 6, 7]]] },
	{ name: "report waits for a member of its group held inside a batch that is still open", steps: [[init(2, 9), [1]], [c(1, { g: 9, b: 2 }), []], [hook(6, { g: 9 }), []], [c(6), []], [fl(2, 0, 9), [2, 3, 4, 5]]] },
	{ name: "two batches of one group, the report waits for the blocked one", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2, 9), [3]], [init(3, 9), [4]], [c(1, { g: 9, b: 2 }), []], [c(5, { g: 9, b: 3 }), []], [fl(2, 0, 9), []], [fl(3, 0, 9), [6, 8]], [hook(6, { g: 9 }), []], [fl(1), [2, 5, 7, 9, 10]]] },
	{ name: "report closing its batch waits for a member of its group held outside the batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2, 9), [3]], [c(1, { g: 9 }), []], [c(2, { g: 9, b: 2 }), []], [hook(6, { g: 9, b: 2 }), []], [c(7), [7]], [fl(1), [2, 4, 5, 6, 8]]] },
	{ name: "batch without a group holds its members and does not delay a report", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2, { g: 9 }), [3]], [hook(3, { g: 9 }), [4]], [fl(1), [2, 5]]] },
	{ name: "member arriving after the closer of its never-opened batch", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(1, { b: 5 }), []], [fl(5), []], [c(2, { b: 5 }), [5]], [c(1, { b: 5 }), []], [fl(1), [2, 3, 4, 6, 7]]] },
	{ name: "member arriving after its batch closed runs at once", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(1), [2, 3]], [c(1, { b: 1 }), [4]]] },
	{ name: "late closer waits behind the batch holding its own door", steps: [[init(1), [1]], [c(1, { b: 1 }), []], [fl(5, 1), []], [fl(1), [2, 3, 4]]] },
	{ name: "cap reached by members, late member and late closer run at once", cap: 2, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [c(2, { b: 1 }), [2, 3]], [c(1, { b: 1 }), [4]], [fl(1), [5]]] },
	{ name: "cap skips an older empty batch, which stays open and holds its later members", cap: 2, steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 2 }), []], [c(2, { b: 2 }), [3, 4]], [c(3, { b: 1 }), []], [fl(1), [5, 6]], [fl(2), [7]]] },
	{ name: "cap forces a batch together with an open parent created later", cap: 3, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2), [3]], [c(2, { b: 2 }), []], [c(2, { b: 1 }), [2, 4, 5]], [c(3, { b: 2 }), [6]], [fl(2), [7]], [fl(1), [8]]] },
	{ name: "cap forced batch, held late member and late closer waiting for it", cap: 4, steps: [[init(1), [1]], [init(2), [2]], [c(1, { b: 1 }), []], [c(5, { b: 2 }), []], [c(2, { b: 1 }), []], [c(3, { b: 1 }), [3, 5, 6]], [c(5, { b: 1 }), []], [fl(1), []], [c(9), [9]], [fl(2), [4, 7, 8, 10]]] },
	{ name: "cap forced batch, held late member and late report waiting for it", cap: 4, steps: [[init(1, 9), [1]], [init(2), [2]], [c(1, { g: 9, b: 1 }), []], [c(5, { b: 2 }), []], [c(2, { g: 9, b: 1 }), []], [c(3, { g: 9, b: 1 }), [3, 5, 6]], [c(5, { g: 9, b: 1 }), []], [hook(8, { g: 9, b: 1 }), []], [c(9), [9]], [fl(2), [4, 7, 8, 10]]] },
	{ name: "cap counts the closers of finished batches blocked behind an open one", cap: 7, steps: [[init(1), [1]], [c(1, { b: 1 }), []], [init(2), [3]], [c(1, { b: 2 }), []], [fl(2), []], [init(3), [6]], [c(1, { b: 3 }), []], [fl(3), []], [init(4), [9]], [c(1, { b: 4 }), []], [fl(4), [2, 4, 5, 7, 8, 10, 11]], [fl(1), [12]]] },
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
