import { CaptureErr, captureErrTypes } from './capture'
import captures from './captures'
import indicator, { IndicatorEntry } from './indicator'
import { requestTimeout, id } from './params'
import { AbortTimer } from './lib'

export type ScopeSet = {
    id: number,
    type: keyof typeof newScope,
    opt: any,
}


export class Hook {
    private res: (value: Response) => void
    private rej: (reason: CaptureErr) => void
    private promise: Promise<Response>
    private scopeStack: Array<Scope> = []
    private abortTimer: AbortTimer | null = null
    constructor(private params: { hookId: number, body: any, event?: Event, scopeQueue: Array<ScopeSet>, indicator: Array<IndicatorEntry> }) {
        this.promise = new Promise((res, rej) => {
            this.res = res
            this.rej = rej
        })
        if (!this.params.scopeQueue || this.params.scopeQueue.length == 0) {
            this.params.scopeQueue = [{ id: -1, type: "free", opt: undefined }]
        }
    }
    wait() {
        return this.promise
    }
    nextScope() {
        return this.params.scopeQueue.shift()
    }
    stackScope(scope: Scope) {
        this.scopeStack.unshift(scope)
    }
    execute() {
        let target: Element | null = null
        if (this.params.event?.target) {
            target = this.params.event.target as Element
        }
        const indicatorId = indicator.start(target, this.params.indicator)
        this.abortTimer = new AbortTimer(requestTimeout)
        fetch(`/d00r/${id}/${this.params.hookId}/`, {
            method: "POST",
            signal: this.abortTimer.signal,
            ...this.params.body,
        }).then(r => {
            if (r.ok) {
                /*
                 const after = r.headers.get("D00r-After")
                 if (after) {
                     const [name, arg] = JSON.parse(after)
                     // this.afterHook(name, arg)
                 } */
                this.ok(r)
                return
            }
            if (r.status === 401 || r.status === 410) {
                this.rej(new CaptureErr(captureErrTypes.stale, r))
            } else if (r.status === 400) {
                this.rej(new CaptureErr(captureErrTypes.format))
            } else if (r.status === 403) {
                this.rej(new CaptureErr(captureErrTypes.done))
            } else if (r.status >= 500 && r.status < 600) {
                this.rej(new CaptureErr(captureErrTypes.server, r))
            } else {
                this.rej(new CaptureErr(captureErrTypes.other, r))
            }
        }).catch(e => {
            if (this.abortTimer!.status == "aborted") {
                this.rej(new CaptureErr(captureErrTypes.blocked))
                return
            }
            this.err(new CaptureErr(captureErrTypes.network, e))
        }).finally(() => {
            this.abortTimer!.abort()
            indicator.end(indicatorId)
        })
    }
    private done() {
        this.scopeStack.forEach(s => s.done(this))
    }
    cancel() {
        if (this.abortTimer) {
            this.abortTimer.abort()
            return
        }
        if (this.params.event) {
            this.params.event.preventDefault()
            this.params.event.stopPropagation()
        }
        this.err(new CaptureErr(captureErrTypes.blocked))
    }
    private ok(r: Response) {
        this.res(r)
        this.done()
    }
    private err(r: CaptureErr) {
        this.rej(r)
        this.done()
    }
}



class Runtime {
    private scopes = new Map<number, Scope>()
    constructor() {

    }
    public done(id: number) {
        this.scopes.delete(id)
    }
    public submit(hook: Hook) {
        const set = hook.nextScope()!
        this.next(hook, set)
    }
    public next(hook: Hook, set: ScopeSet) {
        let scope = this.scopes.get(set.id)
        if (!scope) {
            scope = newScope[set.type](this, set.id)
            this.scopes.set(set.id, scope)
        }
        scope.submit(hook, set.opt)
    }
}


abstract class Scope {
    private counter = 0
    constructor(protected runtime: Runtime, protected id: number) {

    }
    protected get size() {
        return this.counter
    }
    public done(fetch: Hook) {
        this.counter -= 1
        this.complete(fetch)
        if (this.counter > 0) {
            return
        }
        this.runtime.done(this.id)
    }
    public submit(hook: Hook, opt: any) {
        hook.stackScope(this)
        this.counter += 1
        this.process(hook, opt)
    }
    protected promote(hook: Hook) {
        const next = hook.nextScope()
        if (!next) {
            hook.execute()
            return
        }
        this.runtime.next(hook, next)
    }
    protected abstract process(hook: Hook, opt: any): void
    protected abstract complete(hook: Hook): void


}


const newScope = {
    "debounce": (runtime: Runtime, id: number) => new DebounceScope(runtime, id),
    "blocking": (runtime: Runtime, id: number) => new BlockingScope(runtime, id),
    "serial": (runtime: Runtime, id: number) => new SerialScope(runtime, id),
    "frame": (runtime: Runtime, id: number) => new FrameScope(runtime, id),
    "latest": (runtime: Runtime, id: number) => new LatestScope(runtime, id),
    "free": (runtime: Runtime, id: number) => new FreeScope(runtime, id),


} as const;

class DebounceScope extends Scope {
    private durationTimer: any = null
    private limitTimer: any = null
    private hook: Hook | null = null
    private launch() {
        this.clearTimeouts()
        this.promote(this.hook!)
        this.hook = null
    }
    private clearTimeouts() {
        clearTimeout(this.durationTimer)
        clearTimeout(this.limitTimer)
        this.durationTimer = null
        this.limitTimer = null
    }
    private resetDuration(duration: number) {
        if (this.durationTimer) {
            clearTimeout(this.durationTimer)
        }
        this.durationTimer = setTimeout(() => {
            this.launch()
        }, duration)
    }
    private resetLimit(limit: number) {
        if (this.limitTimer != null) {
            return
        }
        this.limitTimer = setTimeout(() => {
            this.launch()
        }, limit)
    }
    protected complete(hook: Hook): void {
        if (this.hook !== hook) {
            return
        }
        this.hook = null
        this.clearTimeouts()
    }
    protected process(hook: Hook, opt: any): void {
        const [duration, limit] = opt
        if (this.hook) {
            this.hook.cancel()
        }
        this.hook = hook
        this.resetDuration(duration)
        this.resetLimit(limit)
    }
}

class SerialScope extends Scope {
    private queue: Array<Hook> = []
    protected complete(hook: Hook): void {
        if (this.queue[0] === hook) {
            this.queue.shift()
            const next = this.queue[0]
            if (next) {
                this.promote(next)
            }
            return
        }
        this.queue = this.queue.filter(h => h !== hook)
    }
    protected process(hook: Hook, _opt: any): void {
        this.queue.push(hook)
        if (this.queue.length == 1) {
            this.promote(hook)
        }
    }
}

class BlockingScope extends Scope {
    protected complete(_fetch: Hook): void {

    }
    protected process(hook: Hook, _opt: any): void {
        if (this.size > 1) {
            hook.cancel()
            return
        }
        this.promote(hook)
    }
}

class FrameScope extends Scope {
    private frameHook: Hook | null = null
    protected complete(hook: Hook): void {
        if (!this.frameHook) {
            return
        }
        if (this.frameHook === hook) {
            this.frameHook = null
            return
        }
        if (this.size != 1) {
            return
        }
        this.promote(this.frameHook)
    }
    protected process(hook: Hook, opt: any): void {
        const frame = opt as boolean
        if (this.frameHook) {
            hook.cancel()
            return
        }
        if (frame !== true) {
            this.promote(hook)
            return
        }
        this.frameHook = hook
        if (this.size != 1) {
            return
        }
        this.promote(hook)
    }
}

class LatestScope extends Scope {
    private last: Hook | null = null
    protected complete(hook: Hook): void {
        if (this.last !== hook) {
            return
        }
        this.last = null
    }
    protected process(hook: Hook, _opt: any): void {
        if (this.last) {
            this.last.cancel()
        }
        this.last = hook
        this.promote(hook)
    }
}

class FreeScope extends Scope {
    protected complete(_hook: Hook): void {
    }
    protected process(hook: Hook, _opt: any): void {
        this.promote(hook)
    }
}

const r = new Runtime()
export default function submit(hook: Hook) {
    r.submit(hook)
}
