import ctrl from './controller'
import { Hook } from './scope'
import indicator from './indicator'
import door from './door'

export const captureErrTypes = {
    blocked: "blocked",
    stale: "stale",
    done: "done",
    notFound: "not_found",
    other: "other",
    network: "network",
    format: "format",
    server: "server",
    debounce: "debounce",
    capture: "capture",
} as const

export class CaptureErr extends Error {
    public meta: any
    constructor(public type: string, private opt?: any) {
        let message: string
        switch (type) {
            case captureErrTypes.notFound:
                message = `capture or hook ${opt} not found`
                break
            case captureErrTypes.blocked:
                message = `hook is blocked by another hook`
                break
            case captureErrTypes.stale:
                message = `instance is stopped`
                break
            case captureErrTypes.done:
                message = `hook is done`
                break
            case captureErrTypes.other:
                message = `Other Error: ${opt?.status}`
                break
            case captureErrTypes.network:
                message = opt?.message
                break
            case captureErrTypes.capture:
                message = opt?.message
                break
            case captureErrTypes.server:
                message = `Server Error: ${opt?.status}`
                break
            case captureErrTypes.format:
                message = `body parsing error, bad request`
                break
            case captureErrTypes.debounce:
                message = `Debounced`
                break
            default:
                throw new Error(`unsupported error type: ${type}`)
        }

        const cause = opt instanceof Error ? opt : undefined
        // @ts-expect-error: Error constructor overload not recognized by TS (ES2022 feature)
        super(message, cause ? { cause } : undefined)
    }

    isBlocked() { return this.type === captureErrTypes.blocked; }
    isNotFound() { return this.type === captureErrTypes.notFound; }
    isStale() { return this.type === captureErrTypes.stale; }
    isDone() { return this.type === captureErrTypes.done; }
    isOther() { return this.type === captureErrTypes.other; }
    isNetwork() { return this.type === captureErrTypes.network; }
    isCapture() { return this.type === captureErrTypes.capture; }
    isServer() { return this.type === captureErrTypes.server; }
    isFormat() { return this.type === captureErrTypes.format; }
    isDebounce() { return this.type === captureErrTypes.debounce; }

    status(): number | undefined {
        return this.opt?.status
    }
}

export function capture(name: string, opt: any, arg: any, event: Event | undefined, hook: any): Promise<Response> {
    const [nodeId, hookId, scopeQueue, indicator] = hook
    const h = new Hook({
        nodeId,
        hookId,
        event: event,
        scopeQueue,
        indicator,
    })
    return h.capture(name, opt, arg)
}

export function attach(parent: HTMLElement | DocumentFragment | Document) {
    for (const element of parent.querySelectorAll<HTMLElement>("[data-d00r-capture]")) {
        const capturesList = JSON.parse(element.getAttribute("data-d00r-capture")!)
        element.removeAttribute("data-d00r-capture")
        for (const [event, name, opt, hook] of capturesList) {
            element.addEventListener(event, async (e) => {
                try {
                    await capture(name, opt, e, e, hook)
                } catch (err: any) {
                    if (!(err instanceof CaptureErr)) {
                        console.error("unknown error in capture:", err)
                        return
                    }
                    if (err.isDebounce() || err.isBlocked() || err.isDone()) {
                        return
                    }
                    if (err.isStale()) {
                        ctrl.gone()
                        return
                    }
                    const onErr = hook[4]
                    if (!onErr || onErr.length == 0) {
                        console.error("capture execution error", err)
                        return
                    }
                    const nodeId = hook[0]
                    for (const [type, args] of onErr) {
                        if (type == "indicator") {
                            const [duration, indications] = args
                            const id = indicator.start(element, indications)
                            if (id) {
                                setTimeout(() => indicator.end(id), duration)
                            }
                        }
                        if (type == "call") {
                            const [name, meta] = args
                            err.meta = meta
                            const handler = door.getHandler(nodeId, name)
                            if (!handler) {
                                console.error("error handeling call " + name + " not found")
                                return
                            }
                            try {
                                await handler[0](err)
                            } catch (e) {
                                console.error("error handeling call " + name + " failed", e)
                            }
                        }
                    }
                }
            })
        }
    }
}

