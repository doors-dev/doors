// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import { id, disconnectAfter, ttl, solitairePing } from "./params"
import action from "./calls"
import { ProgressiveDelay, AbortTimer, ReliableTimer } from "./lib"

import doors from "./door"
import { Package, PackageBuilder } from "./package";


const controlBytes = {
    terminator: 0xFF,
    discard: 0xFD,
    content: 0xFC,
}
const signals = {
    ack: 0x00,
    action: 0x01,
    roll: 0x02,
    suspend: 0x03,
    kill: 0x04,
}


class Solitaire {
    top: Card
    private cursor: number = 1
    constructor() {

    }
    private collectedLost = new Set<number>()
    returnLost(lost: Lost) {
        for (const gap of lost) {
            this.collectedLost.delete(gap)
        }
    }
    private getLost(): Lost {
        let lost: Lost = []
        if (!this.top) {
            return lost
        }
        this.top.lost(this.cursor - 1, lost)
        return lost
    }
    isDone(): boolean {
        return !this.top
    }
    collectLost(): Lost {
        return this.getLost().filter((seq) => {
            if (this.collectedLost.has(seq)) {
                return false
            }
            this.collectedLost.add(seq)
            return true
        })
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
        return a.filter(p => !p.isFiller)
    }
    insert(p: Package) {
        // console.log(this.cursor, !!this.top, "ARRIVED", p.start, p.end, [...this.collectedLost])
        for (let seq = p.start; seq <= p.end; seq++) {
            this.collectedLost.delete(seq)
        }
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
            //      console.log("CONSUMED, EXTENING BOTH SIDED");
            this.p = p
            if (this.next) {
                this.next.cover(this.p.end, this)
            }
            return
        }
        if (p.end >= this.p.end) { // && p.start > this.start
            //       console.log("CONSUMED, EXTENDING RIGHT");
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
    signal: "signal",
    header: "header",
    payload: "payload",
} as const
type SyncStatus = typeof connectorStatus[keyof typeof connectorStatus];


class NetworkError extends Error { }

const reports = {
    ok: "ok",
    broken: "broken",
    interrupted: "interrupt",
} as const;
type Report = typeof reports[keyof typeof reports];

type Lost = Array<number>
type Gap = ([number, number] | [number])
type Gaps = Array<Gap>

class Connection {
    private status: SyncStatus = connectorStatus.signal
    private abortTimer: AbortTimer
    private rollTimer: ReliableTimer
    constructor(private ctrl: Controller, private results: Results, private lost: Lost) {
        this.abortTimer = new AbortTimer(solitairePing * 4 / 3)
        this.rollTimer = new ReliableTimer(solitairePing, () => {
            this.ctrl.requestRoll(this)
        })
        this.run()
    }
    private package: PackageBuilder | undefined
    abort() {
        this.abortTimer.abort()
    }
    private acked = false
    private ack() {
        if (this.acked) {
            return
        }
        this.acked = true

    }
    private report(ok: boolean = false) {
        this.abortTimer.cancel()
        this.rollTimer.cancel()
        const report = ok ? reports.ok : this.acked ? reports.interrupted : reports.broken;
        this.ctrl.report(this, report, this.results, this.lost)
    }
    private get gaps(): Gaps {
        const gaps: Gaps = []
        let gap: any
        for (const seq of this.lost) {
            if (!gap) {
                gap = [seq]
                continue
            }
            const prev = gap[1] !== undefined ? gap[1] : gap[0]
            if (seq == prev + 1) {
                gap[1] = seq
                continue
            }
            gaps.push(gap)
            gap = [seq]
        }
        if (gap) {
            gaps.push(gap)
        }
        return gaps
    }

    private async run() {
        try {
            let response: Response
            try {
                response = await fetch("/d00r/" + id, {
                    signal: this.abortTimer.signal,
                    method: "PUT",
                    headers: {
                        Accept: "application/octet-stream",
                        'Content-Type': 'application/json;charset=UTF-8',
                    },
                    body: JSON.stringify({
                        gaps: this.gaps,
                        results: Object.fromEntries(this.results!),
                    }),
                })
            } catch (e) {
                throw new NetworkError()
            }
            if (response.status === 401 || response.status === 410) {
                this.ctrl.kill()
                throw new Error()
            }
            if (!response.ok) {
                throw new Error()
            }
            const reader = response.body!.getReader()
            while (true) {
                let value: Uint8Array
                const result = await reader.read()
                if (result.done) {
                    throw new Error()
                }
                /*
                if (Math.random() > 0.5) {
                    throw new Error()
                } 
                */

                value = result.value
                const done = this.onChunk(value)
                this.ctrl.flush()
                if (done) {
                    reader.cancel()
                    break
                }
            }
            this.report(true)
        } catch (e) {
            this.report()
        }
    }
    private onChunk(data: Uint8Array): boolean {
        if (data.length == 0) {
            return false
        }
        if (this.status == connectorStatus.signal) {
            const signal = data[0]
            switch (signal) {
                case signals.ack:
                    this.ack()
                    if (data.length == 1) {
                        return false
                    }
                    return this.onChunk(data.subarray(1))
                case signals.action:
                    this.status = connectorStatus.header
                    this.package = new PackageBuilder()
                    if (data.length == 1) {
                        return false
                    }
                    return this.onChunk(data.subarray(1))
                case signals.suspend:
                    this.ctrl.suspend()
                    break;
                case signals.kill:
                    this.ctrl.kill()
                    break;
                case signals.roll:
                    break
                default:
                    console.error(new Error("unsupported signal " + signal))
            }
            return true
        }
        if (this.status == connectorStatus.header) {
            for (let i = 0; i < data.length; i++) {
                const byte = data[i]
                if (byte == controlBytes.discard) {
                    this.package = undefined
                    this.status = connectorStatus.signal
                    if (i + 1 == data.length) {
                        return false
                    }
                    return this.onChunk(data.subarray(i + 1))
                }
                if (byte != controlBytes.content && byte != controlBytes.terminator) {
                    continue
                }
                this.package!.appendHeaderData(data.subarray(0, i))
                if (byte == controlBytes.terminator) {
                    this.ctrl.onPackage(this.package!.build())
                    this.package = undefined
                    this.status = connectorStatus.signal
                } else {
                    this.status = connectorStatus.payload
                }
                if (i + 1 == data.length) {
                    return false
                }
                return this.onChunk(data.subarray(i + 1))
            }
            this.package!.appendHeaderData(data)
            return false
        }
        for (let i = 0; i < data.length; i++) {
            const byte = data[i]
            if (byte == controlBytes.discard) {
                this.package = undefined
                this.status = connectorStatus.signal
                if (i + 1 == data.length) {
                    return false
                }
                return this.onChunk(data.subarray(i + 1))
            }
            if (byte != controlBytes.terminator) {
                continue
            }
            this.package!.appendPayloadData(data.subarray(0, i))
            this.ctrl.onPackage(this.package!.build())
            this.package = undefined
            this.status = connectorStatus.signal
            if (i + 1 == data.length) {
                return false
            }
            return this.onChunk(data.subarray(i + 1))
        }
        this.package!.appendPayloadData(data)
        return false
    }
}

type Results = Map<number, [any, undefined] | [undefined, string]>

class Tracker {
    private buffered: Results = new Map()
    process(p: Package) {
        const [ok, err] = action(p.action, p.arg, { payload: p.payload })

        this.buffered.set(p.end, [ok, err?.message])
    }
    return(collected: Results) {
        for (const [seq, entry] of collected.entries()) {
            this.buffered.set(seq, entry)
        }
    }
    collect(): Results {
        const collected = this.buffered
        this.buffered = new Map()
        return collected
    }
    isDone(): boolean {
        return this.buffered.size == 0
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
    deck = new Solitaire()
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
        this.roll()
    }
    private ttlTimer = new ReliableTimer(ttl, () => this.suspend())
    private resetTtl() {
        this.ttlTimer.reset()
    }
    private ready() {
        this.loaded = true
        doors.scan(document)
        if (this.state == state.dead) {
            return
        }
        this.flush()
    }
    onPackage(p: Package) {
        this.deck.insert(p)
    }
    flush() {
        if (!this.loaded) {
            return
        }
        const collection = this.deck.collect()
        for (const p of collection) {
            this.tracker.process(p)
        }
        this.resetTtl()
        if (this.deck.isDone() && this.tracker.isDone()) {
            return
        }
        this.roll()
    }
    private rolling = false
    private roll(delay = false) {
        if (this.state != state.active) {
            return
        }
        if (this.rolling) {
            return
        }
        if (delay) {
            this.rolling = true
            this.delay.wait().then(() => {
                this.rolling = false
                this.connect()
            })
        }
        this.connect()
    }
    private connect() {
        const results = this.tracker.collect()
        const lost = this.deck.collectLost()
        // console.log("CONNECT", lost, results, this.connections.size)
        if (lost.length == 0 && results.size == 0 && this.connections.size > 1) {
            return
        }
        this.connections.add(new Connection(this, results, lost))
    }
    requestRoll(connection: Connection) {
        // console.log("ROLL REQUESTED", this.connections.has(connection), this.connections.size)
        if (!this.connections.has(connection)) {
            return
        }
        if (this.connections.size != 1) {
            return
        }
        this.roll()
    }
    report(connection: Connection, report: Report, results: Results, lost: Lost) {
        this.connections.delete(connection)
        // console.log(result, returned !== undefined, this.connections.size)
        if (lost.length > 0) {
            if (report == reports.ok) {
                setTimeout(() => {
                    this.deck.returnLost(lost)
                }, 0)
            } else {
                this.deck.returnLost(lost)
            }
        }
        if (report == reports.broken) {
            this.tracker.return(results)
        }
        if (report == reports.ok) {
            this.delay.reset()
        }
        if (this.connections.size >= 6 || (this.connections.size != 0 && report == reports.ok)) {
            return
        }
        this.roll(report == reports.broken)
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
        }, disconnectAfter)
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
        this.roll()
    }
    private reloaded = false
    private reload() {
        if (this.reloaded) {
            return
        }
        this.reloaded = true
        location.reload()
    }
    suspend() {
        if (this.state == state.dead) {
            return
        }
        this.state = state.dead
        this.closeConnections();
        ["pointerdown", "pointermove", "pointerup", "scroll", "focus", "keydown", "input"].forEach(event => {
            window.addEventListener(event, () => this.reload(), true);
        });
    }
    kill() {
        this.state = state.dead
        if (!document.hidden) {
            this.reload()
        }
        return
    }
}

const c = new Controller()

export default {
    gone() {
        c.kill()
    }
}
