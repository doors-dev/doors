// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import { HookErr, hookErrKinds } from './capture'
import captures from './captures'
import indicator, { IndicatorEntry } from './indicator'
import { requestTimeout, id, prefix } from './params'
import { AbortTimer } from './lib'
import action, { Action } from './calls'
import { decodePayload } from './package'

export type ScopeSet = [keyof typeof newScope, string, any]


export class Hook {
	private res_: (value: Response) => void
	private rej_: (reason: HookErr) => void
	private promise_: Promise<Response>
	private scopeStack_: Array<Scope> = []
	private fetch_: any = {}
	private scopeQueue_: Array<ScopeSet>
	private indicatorId_: number | undefined = undefined
	private track_: number | undefined = undefined
	constructor(private params_: {
		doorId: number,
		hookId: number,
		event?: Event,
		scopeQueue: Array<ScopeSet>,
		indicator: Array<IndicatorEntry>,
		before: Array<Action>
	}) {
		this.promise_ = new Promise((res, rej) => {
			this.res_ = res
			this.rej_ = rej
		})
		if (!this.params_.scopeQueue || this.params_.scopeQueue.length == 0) {
			this.scopeQueue_ = [["free", "", undefined]]
		} else {
			this.scopeQueue_ = [...this.params_.scopeQueue]
		}

	}
	capture(name: string, opt: any, arg: any) {
		const captureFunction = captures[name]
		if (!captureFunction) {
			this.err(new HookErr(hookErrKinds.capture, new Error("capture " + name + " not found")))
			return this.promise_
		}
		try {
			this.fetch_ = captureFunction(arg, opt)
			if (this.fetch_ === undefined) {
				this.rej_(new HookErr(hookErrKinds.canceled))
				return this.promise_
			}
		} catch (e) {
			this.rej_(new HookErr(hookErrKinds.capture, e))
			return this.promise_
		}
		runtime.submitHook(this)
		return this.promise_
	}
	nextScope() {
		return this.scopeQueue_.shift()
	}
	stackScope(scope: Scope) {
		this.scopeStack_.unshift(scope)
	}
	private async actions(actions: Array<Action>) {
		for (const [name, arg, payload] of actions) {
			const [_, err] = action(name, arg, { element: this.params_.event?.target as any, payload: await decodePayload(payload) })
			if (err) {
				console.error("hook action err", err)
			}
		}
	}
	private abortTimer_: AbortTimer | undefined
	execute() {
		let target: Element | null = null
		if (this.params_.event?.currentTarget) {
			target = this.params_.event.currentTarget as Element
		}
		this.indicatorId_ = indicator.start(target, this.params_.indicator)
		this.abortTimer_ = new AbortTimer(requestTimeout)
		this.track_ = runtime.hookRegister(this)
		const track = this.track_
		this.actions(this.params_.before).then(() => {
			fetch(`${prefix}/h/${id}/${this.params_.doorId}/${this.params_.hookId}?t=${track}`, {
				method: "POST",
				signal: this.abortTimer_!.signal,
				...this.fetch_,
			}).then(r => {
				this.abortTimer_!.cancel()
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
				if (this.abortTimer_!.status == "aborted") {
					runtime.hookErr(track, new HookErr(hookErrKinds.canceled))
					return
				}
				if (this.abortTimer_!.status == "running") {
					this.abortTimer_!.cancel()
				}
				runtime.hookErr(track, new HookErr(hookErrKinds.network, e))
			})
		})
	}
	private done() {
		this.scopeStack_.forEach(s => s.done(this))
	}
	cancel() {
		if (this.response_) {
			this.response_ = undefined
			runtime.hookErr(this.track_!, new HookErr(hookErrKinds.canceled))
			return
		}
		if (this.abortTimer_) {
			this.abortTimer_.abort()
			return
		}
		if (this.params_.event) {
			this.params_.event.preventDefault()
			this.params_.event.stopPropagation()
		}
		this.err(new HookErr(hookErrKinds.canceled))
	}

	private response_: Response | undefined = undefined
	private reported_ = false
	report(): boolean {
		this.reported_ = true
		if (this.response_) {
			return this.ok(this.response_)
		}
		return false
	}
	ok(r: Response): boolean {
		if (!this.reported_) {
			this.response_ = r
			return false
		}
		const afterActions: Array<Action> = []
		const after = r.headers.get("D0-After")
		if (after) {
			afterActions.push(...JSON.parse(after))
		}
		this.actions(afterActions).then(() => {
			indicator.end(this.indicatorId_)
			this.done()
			this.res_(r)
		})
		return true
	}
	err(r: HookErr) {
		this.rej_(r)
		this.done()
		indicator.end(this.indicatorId_)
	}
}



class Runtime {
	private scopes_ = new Map<string, Scope>()
	private hooks_ = new Map<number, Hook>()
	private track_ = 0
	constructor() {

	}
	public hookIsRegistered(track: number): boolean {
		return this.hooks_.has(track)
	}
	public hookRegister(hook: Hook): number {
		this.track_ += 1
		this.hooks_.set(this.track_, hook)
		return this.track_
	}
	public hookErr(track: number, err: HookErr) {
		this.hooks_.get(track)!.err(err)
		this.hooks_.delete(track)
	}
	public hookOk(track: number, r: Response) {
		if (this.hooks_.get(track)!.ok(r)) {
			this.hooks_.delete(track)
		}
	}
	public hookReport(track: number) {
		if (!this.hookIsRegistered(track)) {
			return
		}
		if (this.hooks_.get(track)!.report()) {
			this.hooks_.delete(track)
		}
	}
	public scopeDone(id: string) {
		this.scopes_.delete(id)
	}
	public submitHook(hook: Hook) {
		const set = hook.nextScope()!
		this.nextScope(hook, set)
	}
	public nextScope(hook: Hook, set: ScopeSet) {
		const [type, id, opt] = set
		let scope = this.scopes_.get(id)
		if (!scope) {
			scope = newScope[type](this, id)
			this.scopes_.set(id, scope)
		}
		scope.submit(hook, opt)
	}
}


abstract class Scope {
	private counter_ = 0
	constructor(protected runtime_: Runtime, protected id_: string) {

	}
	protected get size() {
		return this.counter_
	}
	public done(fetch: Hook) {
		this.counter_ -= 1
		this.complete(fetch)
		if (this.counter_ > 0) {
			return
		}
		this.runtime_.scopeDone(this.id_)
	}
	public submit(hook: Hook, opt: any) {
		hook.stackScope(this)
		this.counter_ += 1
		this.process(hook, opt)
	}
	protected promote(hook: Hook) {
		const next = hook.nextScope()
		if (!next) {
			hook.execute()
			return
		}
		this.runtime_.nextScope(hook, next)
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
	"latest": (runtime: Runtime, id: string) => new LatestScope(runtime, id),
} as const;

class DebounceScope extends Scope {
	private durationTimer_: any = null
	private limitTimer_: any = null
	private hook_: Hook | null = null
	private launch() {
		this.clearTimeouts()
		const hook = this.hook_!
		this.hook_ = null
		this.promote(hook)
	}
	private clearTimeouts() {
		clearTimeout(this.durationTimer_)
		clearTimeout(this.limitTimer_)
		this.durationTimer_ = null
		this.limitTimer_ = null
	}
	private resetDuration(duration: number) {
		if (this.durationTimer_) {
			clearTimeout(this.durationTimer_)
		}
		this.durationTimer_ = setTimeout(() => {
			this.launch()
		}, duration)
	}
	private resetLimit(limit: number) {
		if (this.limitTimer_ != null) {
			return
		}
		if (limit == 0) {
			return
		}
		this.limitTimer_ = setTimeout(() => {
			this.launch()
		}, limit)
	}
	protected complete(hook: Hook): void {
		if (this.hook_ !== hook) {
			return
		}
		this.hook_ = null
		this.clearTimeouts()
	}
	protected process(hook: Hook, opt: any): void {
		const [duration, limit] = opt
		const oldHook = this.hook_
		this.hook_ = hook
		if (oldHook) {
			oldHook.cancel()
		}
		this.resetDuration(duration)
		this.resetLimit(limit)
	}
}

class SerialScope extends Scope {
	private queue_: Array<Hook> = []
	protected complete(hook: Hook): void {
		if (this.queue_[0] === hook) {
			this.queue_.shift()
			const next = this.queue_[0]
			if (next) {
				this.promote(next)
			}
			return
		}
		this.queue_ = this.queue_.filter(h => h !== hook)
	}
	protected process(hook: Hook, _opt: any): void {
		this.queue_.push(hook)
		if (this.queue_.length == 1) {
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
	private groupId_: number = 0

	protected complete(_hook: Hook): void {
	}

	protected process(hook: Hook, opt: any): void {
		const id = opt as number
		if (this.size != 1 && this.groupId_ != id) {
			hook.cancel()
			return
		}
		this.groupId_ = id
		this.promote(hook)
	}
}

class FrameScope extends Scope {
	private frameHook_: Hook | null = null
	protected complete(hook: Hook): void {
		if (!this.frameHook_) {
			return
		}
		if (this.frameHook_ == hook) {
			this.frameHook_ = null
			return
		}
		if (this.size != 1) {
			return
		}
		this.promote(this.frameHook_)

	}
	protected process(hook: Hook, opt: any): void {
		const frame = opt as boolean
		if (this.frameHook_) {
			hook.cancel()
			return
		}
		if (frame !== true) {
			this.promote(hook)
			return
		}
		this.frameHook_ = hook
		if (this.size != 1) {
			return
		}
		this.promote(hook)
	}
}

class LatestScope extends Scope {
	private last_: Hook | null = null
	protected complete(hook: Hook): void {
		if (this.last_ !== hook) {
			return
		}
		this.last_ = null
	}
	protected process(hook: Hook, _opt: any): void {
		if (this.last_) {
			this.last_.cancel()
		}
		this.last_ = hook
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
