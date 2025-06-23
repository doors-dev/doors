import { id, ttl } from './params'
import { randDelay, fetchOptJson } from './lib'
import { attach } from './capture'
import calls from "./calls"



class Link {
    private ready = new Map<number, [number] | [number, string]>()
    private res: (() => void) | null = null
    private last = 0
    private loadedFlag = false
    private buffered: [number, string, any[]][] = []

    loaded() {
        this.loadedFlag = true
        while (true) {
            const next = this.buffered.shift()
            if (!next) break
            this.call(next)
        }
    }

    call([seq, name, args]: [number, string, any[]]) {
        if (!this.loadedFlag) {
            this.buffered.push([seq, name, args])
            return
        }
        if (seq <= this.last) return

        this.last = seq
        const fn = (calls as any)[name]
        if (!fn) {
            this.respond(seq, `callable [${name}] not found`)
            return
        }

        try {
            const result = fn(...args)
            if (result === undefined || !(result instanceof Promise)) {
                this.respond(seq)
                return
            }
            result.then(() => this.respond(seq)).catch(e => this.respond(seq, e.message))
        } catch (e: any) {
            this.respond(seq, e.message)
        }
    }

    private respond(seq: number, error?: string) {
        const entry: [number] | [number, string] = error !== undefined ? [seq, error] : [seq]

        this.ready.set(seq, entry)
        if (this.res) {
            this.res()
        }
    }

    async collect() {
        if (this.ready.size === 0) {
            await new Promise<void>(resolve => this.res = resolve)
            this.res = null
        }
        return { ack: this.last, ready: [...this.ready.values()] }
    }

    clear(collected: { ready: [number, string?][] }) {
        for (const [seq] of collected.ready) {
            this.ready.delete(seq)
        }
    }
}


class Pusher {
    private status: "stopped" | "to_start" | "started" | "to_stop" = "stopped"
    private pkg: any = null

    constructor(private manager: Manager, private link: Link) { }

    start() {
        if (this.status === "stopped") {
            this.status = "to_start"
            this.cycle()
            return
        }
        if (this.status === "to_stop") {
            this.status = "started"
        }
    }

    stop() {
        if (this.status === "to_start") {
            this.status = "stopped"
            return
        }
        if (this.status === "started") {
            this.status = "to_stop"
        }
    }

    private async cycle() {
        if (this.status === "to_stop") {
            this.status = "stopped"
            this.pkg = null
            return
        }

        if (this.status === "to_start") {
            this.status = "started"
            this.pkg = null
        }

        if (this.status === "started" && this.pkg === null) {
            this.pkg = await this.link.collect()
            this.cycle()
            return
        }

        const response = await fetch(`/d00r/${id}`, {
            method: "PUT",
            ...fetchOptJson(this.pkg)
        })

        if (response.ok) {
            this.link.clear(this.pkg)
            this.pkg = null
            this.cycle()
            return
        }

        this.pkg = null
        if (response.status === 401 || response.status === 410) {
            this.manager.gone()
            this.cycle()
            return
        }

        await randDelay()
        this.cycle()
    }
}


class Source {
    private source: EventSource | null = null

    constructor(private manager: Manager, private link: Link) { }

    activate() {
        if (this.source) return
        this.source = new EventSource(`/d00r/${id}`)
        this.source.onerror = async () => {
            this.deactivate()
            await randDelay()
            this.activate()
        }
        this.source.addEventListener("suspend", () => this.manager.suspended())
        this.source.addEventListener("unauthorized", () => this.manager.gone())
        this.source.addEventListener("gone", () => this.manager.gone())
        this.source.addEventListener("call", (event: any) => {
            this.manager.touched()
            if (event.data === "") return
            const call = JSON.parse(event.data)
            this.link.call(call)
        })
    }

    deactivate() {
        if (!this.source) return
        this.source.close()
        this.source = null
    }
}


class Manager {
    private visible = true
    private status: "active" | "suspended" | "gone" | "unloaded" | undefined = undefined
    private timer: any = null

    private readonly source: Source
    private readonly pusher: Pusher

    constructor(private link: Link) {
        this.source = new Source(this, link)
        this.pusher = new Pusher(this, link)
    }

    init() {
        window.addEventListener("beforeunload", () => this.unloaded())
        document.addEventListener("visibilitychange", () => this.setVisible(!document.hidden))

        this.status = "active"
        this.source.activate()
        this.pusher.start()
        this.resetTimer()

        document.addEventListener("DOMContentLoaded", () => {
            attach(document)
            this.link.loaded()
        })
    }

    private setStatus(newStatus: "active" | "suspended" | "gone" | "unloaded") {
        if (newStatus === this.status) return

        if (newStatus === "unloaded" && (this.status === "suspended" || this.status === "gone")) {
            return
        }

        if (newStatus === "active") {
            this.resetTimer()
            this.source.activate()
            this.pusher.start()
            return
        }

        clearTimeout(this.timer)
        this.source.deactivate()
        this.pusher.stop()

        if (newStatus === "gone" && this.visible) {
            location.reload()
        }

        this.status = newStatus
    }

    private resetTimer() {
        clearTimeout(this.timer)
        this.timer = setTimeout(() => this.gone(), ttl * 1000)
    }

    touched() {
        this.resetTimer()
    }

    suspended() {
        this.setStatus("suspended")
    }

    gone() {
        this.setStatus("gone")
    }

    unloaded() {
        this.setStatus("unloaded")
    }

    setVisible(v: boolean) {
        this.visible = v

        if (!this.visible || this.status === "active") return

        if (this.status === "unloaded") {
            this.setStatus("active")
            return
        }

        location.reload()
    }
}

const link = new Link()
const manager = new Manager(link)
manager.init()

export default manager

