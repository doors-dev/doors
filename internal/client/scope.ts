// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import { HookErr, hookErrKinds } from './capture'
import captures from './captures'
import indicator, { IndicatorEntry } from './indicator'
import { requestTimeout, id } from './params'
import { AbortTimer } from './lib'
import action, { Action } from './calls'

export type ScopeSet = [keyof typeof newScope, string, any]


export class Hook {
    private res: (value: Response) => void
    private rej: (reason: HookErr) => void
    private promise: Promise<Response>
    private scopeStack: Array<Scope> = []
    private fetch: any = {}
    private scopeQueue: Array<ScopeSet>
    private indiciatorId: number | undefined = undefined
    constructor(private params: {
        doorId: number,
        hookId: number,
        event?: Event,
        scopeQueue: Array<ScopeSet>,
        indicator: Array<IndicatorEntry>,
        before: Array<Action>
    }) {
        this.promise = new Promise((res, rej) => {
            this.res = res
            this.rej = rej
        })
        if (!this.params.scopeQueue || this.params.scopeQueue.length == 0) {
            this.scopeQueue = [["free", "", undefined]]
        } else {
            this.scopeQueue = [...this.params.scopeQueue]
        }

    }
    capture(name: string, opt: any, arg: any) {
        const captureFunction = captures[name]
        if (!captureFunction) {
            this.err(new HookErr(hookErrKinds.capture, new Error("capture " + name + " not found")))
            return this.promise
        }
        try {
            this.fetch = captureFunction(arg, opt)
            if (this.fetch === undefined) {
                this.rej(new HookErr(hookErrKinds.canceled))
                return this.promise
            }
        } catch (e) {
            this.rej(new HookErr(hookErrKinds.capture, e))
            return this.promise
        }
        runtime.submitHook(this)
        return this.promise
    }
    nextScope() {
        return this.scopeQueue.shift()
    }
    stackScope(scope: Scope) {
        this.scopeStack.unshift(scope)
    }
    private actions(actions: Array<Action>) {
        for (const [name, arg] of actions) {
            const [_, err] = action(name, arg, { element: this.params.event?.target as any, now: true })
            if (err) {
                console.error("after hookaction err", err)
            }
        }
    }
    private launched = false
    execute() {
        let target: Element | null = null
        if (this.params.event?.target) {
            target = this.params.event.target as Element
        }
        this.actions(this.params.before)
        this.indiciatorId = indicator.start(target, this.params.indicator)
        const abortTimer = new AbortTimer(requestTimeout)
        const track = runtime.hookRegister(this)
        this.launched = true
        fetch(`/d00r/${id}/${this.params.doorId}/${this.params.hookId}?t=${track}`, {
            method: "POST",
            signal: abortTimer.signal,
            ...this.fetch,
        }).then(r => {
            abortTimer!.cancel()
            if (r.ok) {
                runtime.hookOk(track, r)
                return
            }
            if (r.status === 401 || r.status === 410) {
                runtime.hookErr(track, new HookErr(hookErrKinds.unauthorized, r))
            } else if (r.status === 400) {
                runtime.hookErr(track, new HookErr(hookErrKinds.bad_request))
            } else if (r.status === 404) {
                runtime.hookErr(track, new HookErr(hookErrKinds.not_found))
            } else if (r.status >= 500 && r.status < 600) {
                runtime.hookErr(track, new HookErr(hookErrKinds.server, r))
            } else {
                runtime.hookErr(track, new HookErr(hookErrKinds.other, r))
            }
        }).catch(e => {
            abortTimer!.cancel()
            runtime.hookErr(track, new HookErr(hookErrKinds.network, e))
        })
    }
    private done() {
        this.scopeStack.forEach(s => s.done(this))
    }
    cancel() {
        if (this.launched) {
            return
        }
        if (this.params.event) {
            this.params.event.preventDefault()
            this.params.event.stopPropagation()
        }
        this.err(new HookErr(hookErrKinds.canceled))
    }

    private response: Response | undefined = undefined
    private reported = false
    report(): boolean {
        this.reported = true
        if (this.response) {
            return this.ok(this.response)
        }
        return false
    }
    ok(r: Response): boolean {
        if (!this.reported) {
            this.response = r
            return false
        }
        const after = r.headers.get("D00r-After")
        if (after) {
            this.actions(JSON.parse(after))
        }
        indicator.end(this.indiciatorId)
        this.done()
        this.res(r)
        return true
    }
    err(r: HookErr) {
        this.rej(r)
        this.done()
        indicator.end(this.indiciatorId)
    }
}



class Runtime {
    private scopes = new Map<string, Scope>()
    private hooks = new Map<number, Hook>()
    private track = 0
    constructor() {

    }
    public hookIsRegistered(track: number): boolean {
        return this.hooks.has(track)
    }
    public hookRegister(hook: Hook): number {
        this.track += 1
        this.hooks.set(this.track, hook)
        return this.track
    }
    public hookErr(track: number, err: HookErr) {
        this.hooks.get(track)!.err(err)
        this.hooks.delete(track)
    }
    public hookOk(track: number, r: Response) {
        if (this.hooks.get(track)!.ok(r)) {
            this.hooks.delete(track)
        }
    }
    public hookReport(track: number) {
        if (!this.hookIsRegistered(track)) {
            return
        }
        if (this.hooks.get(track)!.report()) {
            this.hooks.delete(track)
        }
    }
    public scopeDone(id: string) {
        this.scopes.delete(id)
    }
    public submitHook(hook: Hook) {
        const set = hook.nextScope()!
        this.nextScope(hook, set)
    }
    public nextScope(hook: Hook, set: ScopeSet) {
        const [type, id, opt] = set
        let scope = this.scopes.get(id)
        if (!scope) {
            scope = newScope[type](this, id)
            this.scopes.set(id, scope)
        }
        scope.submit(hook, opt)
    }
}


abstract class Scope {
    private counter = 0
    constructor(protected runtime: Runtime, protected id: string) {

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
        this.runtime.scopeDone(this.id)
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
        this.runtime.nextScope(hook, next)
    }
    protected abstract process(hook: Hook, opt: any): void
    protected abstract complete(hook: Hook): void


}


const newScope = {
    "debounce": (runtime: Runtime, id: string) => new DebounceScope(runtime, id),
    "blocking": (runtime: Runtime, id: string) => new BlockingScope(runtime, id),
    "concurrent": (runtime: Runtime, id: string) => new ConcurrentScope(runtime, id),
    "serial": (runtime: Runtime, id: string) => new SerialScope(runtime, id),
    "frame": (runtime: Runtime, id: string) => new FrameScope(runtime, id),
    "free": (runtime: Runtime, id: string) => new FreeScope(runtime, id),
} as const;

class DebounceScope extends Scope {
    private durationTimer: any = null
    private limitTimer: any = null
    private hook: Hook | null = null
    private launch() {
        this.clearTimeouts()
        const hook = this.hook!
        this.hook = null
        this.promote(hook)
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
        if (limit == 0) {
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
        const oldHook = this.hook
        this.hook = hook
        if (oldHook) {
            oldHook.cancel()
        }
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

class ConcurrentScope extends Scope {
    private groupId: number = 0

    protected complete(_hook: Hook): void {
    }

    protected process(hook: Hook, opt: any): void {
        const id = opt as number
        if (this.size != 1 && this.groupId != id) {
            hook.cancel()
            return
        }
        this.groupId = id
        this.promote(hook)
    }
}

class FrameScope extends Scope {
    private frameHook: Hook | null = null
    protected complete(hook: Hook): void {
        if (!this.frameHook) {
            return
        }
        if (this.frameHook == hook) {
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


class FreeScope extends Scope {
    protected complete(_hook: Hook): void {
    }
    protected process(hook: Hook, _opt: any): void {
        this.promote(hook)
    }
}

const runtime = new Runtime()

export function report(id: number) {
    runtime.hookReport(id)
}
