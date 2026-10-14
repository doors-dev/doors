// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import indicator, { IndicatorEntry } from './indicator'
import { requestTimeout, id, prefix } from './params'
import { AbortTimer, FetchOpt, result } from './lib'
import action, { Action } from './calls'
import { decodePayload } from './package'
import { HookErr, hookErrKinds } from './hook_err'
import controller from './controller'

export type ScopeSet = [keyof typeof newScope, string, any]

export type Fetch = (opt: FetchOpt) => Promise<Response>

export function NewFetch(params: {
	hookId: number,
	event?: Event,
	scopeQueue: Array<ScopeSet>,
	indicator: Array<IndicatorEntry>,
	before: Array<Action>,
	timeout?: number | null
}): Fetch {
	const hook = new Hook(params)
	return hook.fetch
}

class Hook {
	private res: (value: Response) => void
	private rej: (reason: HookErr) => void
	private promise: Promise<Response>
	private scopeStack: Array<Scope> = []
	private fetchOpt: any = {}
	private scopeQueue: Array<ScopeSet>
	private indicatorId: number | undefined = undefined
	private track: number | undefined = undefined
	constructor(private params: {
		hookId: number,
		event?: Event,
		scopeQueue: Array<ScopeSet>,
		indicator: Array<IndicatorEntry>,
		before: Array<Action>,
		timeout?: number | null
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
	fetch = async (opt: FetchOpt): Promise<Response> => {
		this.fetchOpt = opt
		runtime.submitHook(this)
		return this.promise
	}
	nextScope() {
		return this.scopeQueue.shift()
	}
	stackScope(scope: Scope) {
		this.scopeStack.unshift(scope)
	}
	private async actions(actions: Array<Action>) {
		for (const [name, arg, payload] of actions) {
			const [decoded, decodeErr] = await result(async () => await decodePayload(payload))
			if (decodeErr) {
				console.error("hook action payload decode error", decodeErr)
				continue
			}
			const res = action(name, arg, { element: this.params.event?.target as any, payload: decoded })
			const [_, err] = res instanceof Promise ? await res : res
			if (err) {
				console.error("hook action error", err)
			}
		}
	}
	private abortTimer: AbortTimer | undefined
	execute() {
		let target: Element | null = null
		if (this.params.event?.currentTarget) {
			target = this.params.event.currentTarget as Element
		}
		this.indicatorId = indicator.start(target, this.params.indicator)
		this.abortTimer = new AbortTimer(this.params.timeout ?? requestTimeout)
		this.track = runtime.hookRegister(this)
		const track = this.track
		this.actions(this.params.before).then(() => {
			fetch(`${prefix}/h/${id}/${this.params.hookId}?t=${track}`, {
				method: "POST",
				signal: this.abortTimer!.signal,
				...this.fetchOpt,
			}).then(r => {
				this.abortTimer!.cancel()
				if (r.ok) {
					controller.resetDelays()
					runtime.hookOk(track, r)
					return
				}
				if (r.status === 410) {
					runtime.hookErr(track, new HookErr(hookErrKinds.gone, r))
				} else if (r.status === 400) {
					runtime.hookErr(track, new HookErr(hookErrKinds.bad_request))
				} else if (r.status === 404) {
					runtime.hookErr(track, new HookErr(hookErrKinds.not_found))
				} else if (r.status === 412) {
					runtime.hookErr(track, new HookErr(hookErrKinds.canceled))
				} else if (r.status >= 500 && r.status < 600) {
					runtime.hookErr(track, new HookErr(hookErrKinds.server, r))
				} else {
					runtime.hookErr(track, new HookErr(hookErrKinds.other, r))
				}
			}).catch(e => {
				if (this.abortTimer!.status == "aborted") {
					runtime.hookErr(track, new HookErr(hookErrKinds.canceled))
					return
				}
				if (this.abortTimer!.status == "running") {
					this.abortTimer!.cancel()
				}
				runtime.hookErr(track, new HookErr(hookErrKinds.network, e))
			})
		})
	}
	private done() {
		this.scopeStack.forEach(s => s.done(this))
	}
	cancel() {
		if (this.response) {
			this.response = undefined
			runtime.hookErr(this.track!, new HookErr(hookErrKinds.canceled))
			return
		}
		if (this.abortTimer) {
			this.abortTimer.abort()
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
		const afterActions: Array<Action> = []
		const after = r.headers.get("D0-After")
		if (after) {
			afterActions.push(...JSON.parse(after))
		}
		this.actions(afterActions).then(() => {
			indicator.end(this.indicatorId)
			this.done()
			this.res(r)
		})
		return true
	}
	err(r: HookErr) {
		this.rej(r)
		this.done()
		indicator.end(this.indicatorId)
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
	"rate": (runtime: Runtime, id: string) => new RateScope(runtime, id),
	"debounce": (runtime: Runtime, id: string) => new DebounceScope(runtime, id),
	"blocking": (runtime: Runtime, id: string) => new BlockingScope(runtime, id),
	"concurrent": (runtime: Runtime, id: string) => new ConcurrentScope(runtime, id),
	"serial": (runtime: Runtime, id: string) => new SerialScope(runtime, id),
	"frame": (runtime: Runtime, id: string) => new FrameScope(runtime, id),
	"free": (runtime: Runtime, id: string) => new FreeScope(runtime, id),
	"latest": (runtime: Runtime, id: string) => new LatestScope(runtime, id),
} as const;

class RateScope extends Scope {
	private rateTimer: any = null
	private candidate: [Hook, number] | null = null
	private next() {
		this.rateTimer = null
		if (!this.candidate) {
			return
		}
		const [hook, tick] = this.candidate
		this.candidate = null
		this.rateTimer = setTimeout(() => this.next(), tick)
		this.promote(hook)
	}
	protected process(hook: Hook, opt: any): void {
		const tick = opt as number;
		if (this.candidate) {
			this.candidate[0].cancel()
			this.candidate = null
		}
		if (this.rateTimer != null) {
			this.candidate = [hook, tick]
			return
		}
		this.rateTimer = setTimeout(() => this.next(), tick)
		this.promote(hook)
	}
	protected complete(_: Hook): void {
	}

}

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
	private framePromoted = false
	protected complete(hook: Hook): void {
		if (!this.frameHook) {
			return
		}
		if (this.frameHook == hook) {
			this.frameHook = null
			this.framePromoted = false
			return
		}
		if (this.framePromoted) {
			return
		}
		if (this.size != 1) {
			return
		}
		this.framePromoted = true
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
		this.framePromoted = true
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

const runtime = new Runtime()

export function report(id: number) {
	runtime.hookReport(id)
}
