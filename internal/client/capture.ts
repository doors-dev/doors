import ctrl from './controller'
import { Hook } from './scope'
import indicator from './indicator'
import door from './door'

export const captureErrKinds = {
    canceled: "canceled",
    unauthorized: "unauthorized",
    not_found: "not_found",
    other: "other",
    network: "network",
    bad_request: "bad_request",
    server: "server",
    capture: "capture",
} as const

type CaptureErrKind = typeof captureErrKinds[keyof typeof captureErrKinds]

export class CaptureErr extends Error {
    public meta: any
    public status: number | undefined = undefined
    public cause: Error | undefined = undefined
    constructor(public kind: CaptureErrKind, opt?: any) {
        let message: string
        switch (kind) {
            case captureErrKinds.not_found:
                message = `hook not found on server, may be done`
                break
            case captureErrKinds.canceled:
                message = `hook is blocked by scope`
                break
            case captureErrKinds.unauthorized:
                message = `instance is stopped`
                break
            case captureErrKinds.other:
                message = `Other Error: ${opt?.status}`
                break
            case captureErrKinds.network:
                message = opt?.message
                break
            case captureErrKinds.capture:
                message = opt?.message
                break
            case captureErrKinds.server:
                message = `Server Error: ${opt?.status}`
                break
            case captureErrKinds.bad_request:
                message = `body parsing error, bad request`
                break
            default:
                throw new Error(`unsupported error type: ${kind}`)
        }

        const cause = opt instanceof Error ? opt : undefined
        // @ts-expect-error: Error constructor overload not recognized by TS (ES2022 feature)
        super(message, cause ? { cause } : undefined)
        if (opt && opt.status && typeof opt.status == "number") {
            this.status = opt.status
        }
        if (cause) {
            this.cause = cause
        }
    }
    canceled() { return this.kind === captureErrKinds.canceled; }
    notFound() { return this.kind === captureErrKinds.not_found; }
    unauthorized() { return this.kind === captureErrKinds.unauthorized; }
    other() { return this.kind === captureErrKinds.other; }
    network() { return this.kind === captureErrKinds.network; }
    capture() { return this.kind === captureErrKinds.capture; }
    server() { return this.kind === captureErrKinds.server; }
    badRequest() { return this.kind === captureErrKinds.bad_request; }
}

export function capture(name: string, opt: any, arg: any, event: Event | undefined, hook: any): Promise<Response> {
    const [doorId, hookId, scopeQueue, indicator] = hook
    const h = new Hook({
        doorId,
        hookId,
        event: event,
        scopeQueue,
        indicator,
    })
    return h.capture(name, opt, arg)
}

const attr = "data-d00r-capture"
export function attach(parent: Element | DocumentFragment | Document) {
    for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="applied"])`)) {
        const capturesList = JSON.parse(element.getAttribute(attr)!)
        element.setAttribute(attr, "applied")
        for (const [event, name, opt, hook] of capturesList) {
            element.addEventListener(event, async (e) => {
                try {
                    await capture(name, opt, e, e, hook)
                } catch (err: any) {
                    if (!(err instanceof CaptureErr)) {
                        console.error("unknown error in capture:", err)
                        return
                    }
                    if (err.canceled() || err.notFound()) {
                        return
                    }
                    if (err.unauthorized()) {
                        ctrl.gone()
                        return
                    }
                    const onErr = hook[4]
                    if (!onErr || onErr.length == 0) {
                        console.error("capture execution error", err)
                        return
                    }
                    const doorId = hook[0]
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
                            const handler = door.getHandler(doorId, name)
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

