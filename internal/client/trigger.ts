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

import { HookErr, hookErrKinds } from "./hook_err"

const trigger = Symbol()
let active: TriggerMeta | undefined

export type TriggerMeta = {
	promises: Promise<Response>[]
}

export function putTrigger(event: Event): TriggerMeta {
	const meta: TriggerMeta = { promises: [] };
	(event as any)[trigger] = meta
	return meta
}

export function getTrigger(event: Event): TriggerMeta | undefined {
	return (event as any)[trigger] ?? active
}

export function dispatch(target: EventTarget, event: Event): Promise<Response>[] {
	const meta = putTrigger(event)
	const prev = active
	active = meta
	try {
		target.dispatchEvent(event)
	} finally {
		active = prev
	}
	return meta.promises
}

export type EmitReport = {
	ok: number
	canceled: number
	failed: HookErr[]
}

export async function settle(promises: Promise<Response>[]): Promise<EmitReport> {
	const report: EmitReport = { ok: 0, canceled: 0, failed: [] }
	for (const r of await Promise.allSettled(promises)) {
		if (r.status == "fulfilled") {
			report.ok += 1
			continue
		}
		const err = r.reason instanceof HookErr ? r.reason : new HookErr(hookErrKinds.other, r.reason)
		if (err.canceled()) {
			report.canceled += 1
			continue
		}
		report.failed.push(err)
	}
	return report
}

export function refused(target: EventTarget): boolean {
	const el = target instanceof Element ? target : target instanceof Node ? target.parentElement : null
	if (!el) {
		return false
	}
	if (el.matches(":disabled")) {
		return true
	}
	let n: Element | null = el
	while (n) {
		if (n.hasAttribute("inert")) {
			return true
		}
		n = n.parentElement ?? (n.getRootNode() as ShadowRoot).host ?? null
	}
	return false
}
