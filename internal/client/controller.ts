import { id, sleepAfter, ttl, requestTimeout } from "./params"
import calls from "./calls"
import { ProgressiveDelay, AbortTimer } from "./lib"
import { attach } from "./capture"


const decoder = new TextDecoder()

class Package {
    start: number
    end: number
    call: string
    arg: any
    private parts: Array<string> = []
    filler = false
    get seq(): number {
        return this.end
    }
    setHeader(buf: Uint8Array) {
        const arr = JSON.parse(decoder.decode(buf))
        if (arr.length == 1) {
            this.end = arr[0]
            this.start = arr[0]
            this.filler = true
            return
        }
        if (arr.length == 2) {
            [this.end, this.start] = arr
            this.filler = true
        }
        if (arr.length == 3) {
            [this.end, this.call, this.arg] = arr
            this.start = this.end
            return
        }
        [this.end, this.start, this.call, this.arg] = arr
    }
    get payload(): string {
        return this.parts.join("")
    }
    appendData(buf: Uint8Array) {
        if (buf.length == 0) {
            return
        }
        this.parts.push(decoder.decode(buf))
    }
}

type Gaps = Array<[number, number] | [number]>

class Solitaire {
    top: Card
    cursor: number = 1
    constructor() {

    }
    gaps(): Gaps {
        const a: Gaps = []
        if (!this.top) {
            return a
        }
        this.top.gaps(this.cursor - 1, a)
        return a
    }
    collect(): Array<Package> {
        const a: Array<Package> = []
        if (!this.top) {
            return a
        }
        if (this.top.start != this.cursor) {
            return a
        }
        this.top.collect(a, this)
        const tail = a[a.length - 1]
        this.cursor = tail.end + 1
        return a.filter(p => !p.filler)
    }
    insert(p: Package) {
        // console.log(this.cursor, !!this.top, "ARRIVED", p.start, p.end, p.call, p.arg)
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

class Card {
    private next: Card
    constructor(private p: Package) {
    }
    get start(): number {
        return this.p.start
    }
    gaps(end: number, a: Gaps) {
        if (end < this.p.start - 1) {
            const lostStart = end + 1
            const lostEnd = this.p.start - 1
            const value: ([number, number] | [number]) = lostStart == lostEnd ? [lostEnd] : [lostStart, lostEnd]
            a.push(value)
        }
        if (this.next) {
            this.next.gaps(this.p.end, a)
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
        // console.log(this.p.start, this.p.end, "INSERTING", p.start, p.end)
        if (p.end < this.p.start) {
            // console.log("REPLACED");
            if (this.next) {
                this.next.insert(this.p)
            } else {
                this.next = new Card(this.p)
            }
            this.p = p
            return
        }
        if (p.start > this.p.end) {
            // console.log("PASSED");
            if (this.next) {
                this.next.insert(p)
            } else {
                this.next = new Card(p)
            }
            return
        }
        if (p.start == this.p.end) {
            // console.log("CONSUMED, EXTENDING RIGHT FROM TAIL");
            p.start = this.p.start
            this.p = p
            if (this.next) {
                this.next.cover(this.p.end, this)
            }
            return
        }
        // start < end
        if (p.start <= this.p.start && p.end >= this.p.end) {
            // console.log("CONSUMED, EXTENING BOTH SIDED");
            this.p = p
            if (this.next) {
                this.next.cover(this.p.end, this)
            }
            return
        }
        if (p.end >= this.p.end) { // && p.start > this.start
            // console.log("CONSUMED, EXTENDING RIGHT");
            p.start = this.p.start
            this.p = p
            if (this.next) {
                this.next.cover(this.p.end, this)
            }
            return
        }
        // p.end < this.end
        // console.log("ENDEDN BEFORE, EXTENDING RIGHT");
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

const connectorStatus = {
    length: "length",
    signal: "signal",
    header: "header",
    payload: "payload",
} as const
type SyncStatus = typeof connectorStatus[keyof typeof connectorStatus];

class NetworkErr extends Error {
    constructor() {
        super("sync request network error")
    }
}
class RequestErr extends Error {
    constructor(public status: number) {
        super("sync request error, status: " + status)
    }
}

class Connection {
    private status: SyncStatus = connectorStatus.length
    private abortTimer: AbortTimer
    constructor(private ctrl: Controller) {
        this.abortTimer = new AbortTimer(requestTimeout)
        this.run()
    }
    private offset = 0
    private buf = new Uint8Array(4)
    private package: Package | undefined
    abort() {
        this.abortTimer.abort()
    }
    private onChunk(data: Uint8Array) {
        if (data.length == 0) {
            return
        }
        if (this.status == connectorStatus.length) {
            const remainingLength = this.buf.length - this.offset
            const lengthToWrite = Math.min(remainingLength, data.length)
            this.buf.set(data.subarray(0, lengthToWrite), this.offset)
            this.offset += lengthToWrite
            if (this.offset != this.buf.length) {
                return
            }
            this.offset = 0
            const length = new DataView(this.buf.buffer).getUint32(0)
            if (length == 0) {
                this.status = connectorStatus.signal
                if (lengthToWrite == data.length) {
                    return
                }
                return this.onChunk(data.subarray(lengthToWrite))
            }
            this.buf = new Uint8Array(length)
            this.status = connectorStatus.header
            this.package = new Package()
            if (lengthToWrite == data.length) {
                return
            }
            return this.onChunk(data.subarray(lengthToWrite))
        }
        if (this.status == connectorStatus.signal) {
            const signal = data[0]
            this.status = connectorStatus.length
            this.ctrl.signal(signal)
            if (data.length == 1) {
                return
            }
            return this.onChunk(data.subarray(1))
        }
        if (this.status == connectorStatus.header) {
            const remainingLength = this.buf.length - this.offset
            const lengthToWrite = Math.min(remainingLength, data.length)
            this.buf.set(data.subarray(0, lengthToWrite), this.offset)
            this.offset += lengthToWrite
            if (this.offset != this.buf.length) {
                return
            }
            this.offset = 0
            this.package!.setHeader(this.buf)
            this.buf = new Uint8Array(4)
            this.status = connectorStatus.payload
            if (lengthToWrite == data.length) {
                return
            }
            return this.onChunk(data.subarray(lengthToWrite))
        }
        if (this.status == connectorStatus.payload) {
            for (let i = 0; i < data.length; i++) {
                const char = data[i]
                if (char != 0xFF) {
                    continue
                }
                this.package!.appendData(data.subarray(0, i))
                this.ctrl.onPackage(this.package!)
                this.package = undefined
                this.status = connectorStatus.length
                if (i + 1 == data.length) {
                    return
                }
                this.onChunk(data.subarray(i + 1))
                return
            }
            this.package!.appendData(data)
        }
    }
    private done(e?: Error) {
        this.abortTimer.clean()
        this.ctrl.done(this, e)
    }
    private async run() {
        let response: Response
        const results = this.ctrl.tracker.results()
        const gaps = this.ctrl.desk.gaps()
        try {
            response = await fetch("/d00r/" + id, {
                signal: this.abortTimer.signal,
                method: "PUT",
                headers: {
                    Accept: "application/octet-stream",
                    'Content-Type': 'application/json;charset=UTF-8',
                },
                body: JSON.stringify({
                    gaps: gaps,
                    results: Object.fromEntries(results),
                }),
            })
        } catch (e) {
            this.ctrl.tracker.cancel(results)
            if (this.abortTimer.signal.aborted) {
                this.done()
                return
            }
            this.ctrl.done(this, new NetworkErr())
            return
        }
        if (response.status === 401 || response.status === 410) {
            this.ctrl.tracker.cancel(results)
            this.abortTimer.abort()
            this.ctrl.gone()
            return
        }
        if (!response.ok) {
            this.ctrl.tracker.cancel(results)
            this.done(new RequestErr(response.status))
            return
        }
        const reader = response.body!.getReader()
        let confirmed = false
        while (true) {
            let result: ReadableStreamReadResult<Uint8Array>
            try {
                result = await reader.read()
            } catch (e) {
                break
            }
            const { done, value } = result
            if (done) {
                break
            }
            if (value.length == 0) {
                continue
            }
            if (!confirmed) {
                confirmed = true
                this.ctrl.tracker.confirm(results)
            }
            this.onChunk(value)
        }
        if (!confirmed) {
            this.ctrl.tracker.cancel(results)
        }
        this.done()
    }
}

const signals = {
    "connect": 0x00,
    "suspend": 0x01,
    "killed": 0x02,
}

type Results = Map<number, string | null>

class Tracker {
    private buffered: Results = new Map()
    private collected: Results = new Map()
    process(p: Package) {
        const fn = (calls as any)[p.call]
        if (!fn) {
            this.respond(p.seq, `callable [${p.call}] not found`)
            return
        }
        try {
            const result = fn(p.arg, p.payload)
            if (result === undefined || !(result instanceof Promise)) {
                this.respond(p.seq)
                return
            }
            result.then(() => this.respond(p.seq)).catch(e => this.respond(p.seq, e?.message ? e.message : "unkown error"))
        } catch (e: any) {
            this.respond(p.seq, e?.message ? e.message : "unkown error")
        }
    }
    private respond(seq: number, error?: string) {
        this.buffered.set(seq, error === undefined ? null : error)
    }

    confirm(collected: Results) {
        for (const seq of collected.keys()) {
            this.collected.delete(seq)
        }
    }
    cancel(collected: Results) {
        for (const seq of collected.keys()) {
            const value = this.collected.get(seq)!
            this.collected.delete(seq)
            this.buffered.set(seq, value)
        }
    }
    results(): Results {
        const collected = this.buffered
        this.buffered = new Map()
        for (const [key, value] of collected) {
            this.collected.set(key, value)
        }
        return collected
    }
}

const state = {
    dead: "dead",
    sleep: "sleep",
    active: "active",
} as const

type State = typeof state[keyof typeof state]

class Controller {
    private connections = new Set<Connection>()
    private state: State = state.active
    private loaded = false
    private delay = new ProgressiveDelay()
    desk = new Solitaire()
    tracker = new Tracker()
    constructor() {
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
            this.ready()
        })
        this.connect()
        this.resetTtl()
    }
    private ttlTimer: any
    private resetTtl() {
        clearTimeout(this.ttlTimer)
        this.ttlTimer = setTimeout(() => {
            this.suspend()
        }, ttl)
    }
    private ready() {
        this.loaded = true
        attach(document)
        if (this.state == state.dead) {
            return
        }
        this.collect()
    }
    onPackage(p: Package) {
        this.resetTtl()
        this.desk.insert(p)
        if (!this.loaded) {
            return
        }
        this.collect()
    }
    private collect() {
        const collection = this.desk.collect()
        for (const p of collection) {
            this.tracker.process(p)
        }
    }
    private connect() {
        this.connections.add(new Connection(this))
    }
    private sleepTimer: any = null
    private sleep() {
        clearTimeout(this.sleepTimer)
        this.sleepTimer = null
        if (this.state != state.active) {
            return
        }
        this.state = state.sleep
        this.closeConnections()
    }
    private hidden() {
        if (this.state != state.active) {
            return
        }
        if (this.sleepTimer != null) {
            return
        }
        this.sleepTimer = setTimeout(() => {
            this.sleep()
        }, sleepAfter)
    }

    private closeConnections() {
        for (const connection of this.connections) {
            connection.abort()
        }
    }
    private showed() {
        clearTimeout(this.sleepTimer)
        this.sleepTimer = null
        if (this.state == state.dead) {
            this.reload()
            return
        }
        this.delay.reset()
        if (this.state != state.sleep) {
            return
        }
        this.state = state.active
        if (this.connections.size != 0) {
            return
        }
        this.connect()
    }
    private reloaded = false
    private reload() {
        if (this.reloaded) {
            return
        }
        this.reloaded = true
        location.reload()
    }
    private suspend() {
        if (this.state == state.dead) {
            return
        }
        this.state = state.dead
        this.closeConnections();
        ["pointerdown", "pointermove", "pointerup", "scroll", "focus", "keydown", "input"].forEach(event => {
            window.addEventListener(event, () => this.reload(), true);
        });
    }
    signal(signal: number) {
        if (signal == signals.connect && this.state == state.active) {
            this.resetTtl()
            this.connect()
            return
        }
        if (signal == signals.suspend) {
            this.suspend()
            return
        }
        if (signal == signals.killed) {
            this.state = state.dead
            if (!document.hidden) {
                this.reload()
            }
            return
        }
    }
    gone() {
        this.signal(signals.killed)
    }
    done(connection: Connection, e: Error | undefined) {
        this.connections.delete(connection)
        if (this.connections.size != 0 || this.state != state.active) {
            return
        }
        if (e) {
            this.delay.wait().then(() => {
                if (this.connections.size != 0) {
                    return
                }
                this.connect()
            })
            return
        }
        this.connect()
    }
}

const c = new Controller()

export default {
    gone() {
        c.gone()
    }
}
