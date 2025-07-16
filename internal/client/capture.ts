import { id } from './params'
import ctrl from './controller'
import calls from './calls'
import captures from './captures'
import indicator, { IndicatorEntry } from './indicator'


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

class Process {
    private counter = 0
    private wg = new Map<number, Promise<void>>()

    constructor(
        private runtime: Runtime,
        private hookId: string,
        private mode: HookMode,
    ) { }

    add() {
        const id = this.counter++
        let res!: () => void
        this.wg.set(id, new Promise<void>(r => res = r))
        return () => {
            this.wg.delete(id)
            res()
            if (!this.isActive()) {
                this.runtime.removeProcess(this.hookId)
            }
        }
    }

    isActive() {
        return this.wg.size > 0
    }

    suppress(event: any) {
        if (event instanceof Event) {
            event.preventDefault()
            event.stopPropagation()
        }
    }

    private afterHook(name: string, arg: any) {
        try {
            const fn = (calls as any)[name]
            if (!fn) {
                console.error(`after hook callable [${name}] not found`)
                return
            }
            const result = fn(arg)
            if (result && result instanceof Promise) {
                result.then().catch(e => console.error("after hook  err", e))
            }
        } catch (e) {
            console.error("after hook  err", e)
        }
    }

    hook(task: Task, done: () => void) {
        const indicatorId = indicator.start(task.arg?.target, task.hook!.indicator)
        fetch(`/d00r/${id}/${task.hook!.id}`, {
            method: "POST",
            ...task.fetch,
        }).then(r => {
            if (r.ok) {
                const after = r.headers.get("D00r-After")
                if (after) {
                    const [name, arg] = JSON.parse(after)
                    this.afterHook(name, arg)
                }
                task.res(r)
                return
            }
            if (r.status === 401 || r.status === 410) {
                task.rej(new CaptureErr(captureErrTypes.stale, r))
            } else if (r.status === 400) {
                task.rej(new CaptureErr(captureErrTypes.format))
            } else if (r.status === 403) {
                task.rej(new CaptureErr(captureErrTypes.done))
            } else if (r.status >= 500 && r.status < 600) {
                task.rej(new CaptureErr(captureErrTypes.server, r))
            } else {
                task.rej(new CaptureErr(captureErrTypes.other, r))
            }
        }).catch(e => {
            task.rej(new CaptureErr(captureErrTypes.network, e))
        }).finally(() => {
            done()
            indicator.end(indicatorId)
        })
    }

    capture(task: Task): boolean {
        const captureFn = captures[task.name]
        if (!captureFn) {
            task.rej(new CaptureErr(captureErrTypes.notFound, task.name))
            return false
        }
        try {
            task.fetch = captureFn(task.arg, task.opt)
            return true
        } catch (e) {
            task.rej(new CaptureErr(captureErrTypes.capture, e))
            return false
        }
    }

    wait() {
        return Promise.all(this.wg.values())
    }

    submit(task: Task) {
        switch (this.mode.value) {
            case "debounce":
                this.debounceSubmit(task)
                break
            case "frame":
                this.frameSubmit(task)
                break
            case "butter":
                this.butterSubmit(task)
                break
            case "block":
                this.blockSubmit(task)
                break
            default:
                this.defaultSubmit(task)
        }
    }


    private debounceSubmit(task: Task) {
        if (!this.debounceState) {
            this.debounceState = new DebounceState(this, this.runtime)
        }
        this.debounceState.submit(task)
    }

    private frameSubmit(task: Task) {
        if (this.isActive()) {
            this.suppress(task.arg)
            task.rej(new CaptureErr(captureErrTypes.blocked))
            return
        }
        const ok = this.capture(task)
        if (!ok) return

        const promise = this.runtime.frameStart()
        const done = this.add()

        promise.then(() => {
            this.hook(task, () => {
                done()
                this.runtime.frameEnd()
            })
        })
    }

    private butterSubmit(task: Task) {
        const ok = this.capture(task)
        if (!ok) return
        this.hook(task, this.add())
    }

    private blockSubmit(task: Task) {
        if (this.runtime.isFraming() || this.isActive()) {
            this.suppress(task.arg)
            task.rej(new CaptureErr(captureErrTypes.blocked))
            return
        }
        const ok = this.capture(task)
        if (!ok) return
        this.hook(task, this.add())
    }

    private defaultSubmit(task: Task) {
        if (this.runtime.isFraming()) {
            this.suppress(task.arg)
            task.rej(new CaptureErr(captureErrTypes.blocked))
            return
        }
        const ok = this.capture(task)
        if (!ok) return
        this.hook(task, this.add())
    }

    private debounceState?: DebounceState
}

class DebounceState {
    private doneHook: (() => void) | null = null
    private durationTimer: any = null
    private limitTimer: any = null
    private task: Task | null = null

    constructor(private process: Process, private runtime: Runtime) { }

    submit(task: Task) {
        if (!this.doneHook && this.runtime.isFraming()) {
            this.process.suppress(task.arg)
            task.rej(new CaptureErr(captureErrTypes.blocked))
            return
        }
        const ok = this.process.capture(task)
        if (!ok) return

        if (!this.doneHook) {
            this.doneHook = this.process.add()
            this.setLimit(task)
        }
        this.arm(task)
    }

    private arm(task: Task) {
        const [duration] = task.hook!.mode.args
        clearTimeout(this.durationTimer)
        if (this.task) {
            this.task.rej(new CaptureErr(captureErrTypes.debounce))
        }
        this.task = task
        this.durationTimer = setTimeout(() => this.fire(), duration)
    }

    private setLimit(task: Task) {
        const [, limit] = task.hook!.mode.args
        if (limit <= 0) return
        this.limitTimer = setTimeout(() => this.fire(), limit)
    }

    private fire() {
        const done = this.doneHook!
        this.doneHook = null
        clearTimeout(this.limitTimer)
        clearTimeout(this.durationTimer)
        const task = this.task
        this.task = null
        this.process.hook(task!, done)
    }
}

class Runtime {
    private frameCounter = 0
    private processes = new Map<string, Process>()
    private blocked = false

    frameStart() {
        this.frameCounter++
        return Promise.all([...this.processes.values()].map(p => p.wait()))
    }

    frameEnd() {
        this.frameCounter--
    }

    isFraming() {
        return this.frameCounter !== 0
    }

    removeProcess(id: string) {
        this.processes.delete(id)
    }

    submit(task: Task) {
        if (!task.hook) {
            this.submitLocalTask(task)
            return
        }
        let thread = this.processes.get(task.hook.id)
        if (!thread) {
            thread = new Process(this, task.hook.id, task.hook.mode)
            this.processes.set(task.hook.id, thread)
        }
        thread.submit(task)
        if (!thread.isActive()) {
            this.processes.delete(task.hook.id)
        }
    }

    private submitLocalTask(task: Task) {
        if (this.blocked) {
            if (task.arg instanceof Event) {
                task.arg.preventDefault()
                task.arg.stopPropagation()
            }
            task.rej(new CaptureErr(captureErrTypes.blocked))
            return
        }
        const captureFn = captures[task.name]
        if (!captureFn) {
            task.rej(new CaptureErr(captureErrTypes.notFound, task.name))
            return
        }
        try {
            captureFn(task.arg, task.opt)
        } catch (e) {
            task.rej(new CaptureErr(captureErrTypes.capture, e))
        }
    }
}

const runtime = new Runtime()


interface HookMode {
    value: string; args: any[]
}

interface Hook {
    mark: string
    id: string
    mode: HookMode,
    indicator: IndicatorEntry[] | null
}

class Hook {
    public id: string
    public arg: any
    public mark: string
    public mode: { value: string; args: any[] }
    public indicator: IndicatorEntry[] | null
    constructor(raw: any) {
        const [nodeId, hookId, [mode, ...modeArgs], indicator, mark] = raw
        this.mark = mark
        this.id = `${nodeId}/${hookId}`
        this.mode = {
            value: mode,
            args: modeArgs,
        }
        this.indicator = indicator
    }
}

interface Task {
    name: string
    arg: any
    opt: any
    hook: Hook | null
    fetch: any
    res: (value: Response) => void
    rej: (reason: CaptureErr) => void
}


export function capture(name: string, arg: any, opt: any, hook: any): Promise<Response> {
    return new Promise((res, rej) => {
        const task: Task = {
            name,
            arg,
            opt,
            hook: hook ? new Hook(hook) : null,
            fetch: undefined,
            res,
            rej
        }
        runtime.submit(task)
    })
}

export function attach(parent: HTMLElement | DocumentFragment | Document) {
    for (const element of parent.querySelectorAll<HTMLElement>("[data-d00r-capture]")) {
        const capturesList = JSON.parse(element.getAttribute("data-d00r-capture")!)
        element.removeAttribute("data-d00r-capture")

        for (const [event, name, opt, hook] of capturesList) {
            element.addEventListener(event, async (e) => {
                try {
                    await capture(name, e, opt, hook)
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
                    }
                }
            })
        }
    }
}

