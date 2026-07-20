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
import controller from "./controller"

const collector = Symbol()

export function collect(event: Event, promise: Promise<Response>) {
	(event as any)[collector]?.push(promise)
}

export async function dispatch(target: EventTarget, event: Event): Promise<number> {
	await controller.ready
	const promises: Promise<Response>[] = [];
	(event as any)[collector] = promises
	try {
		target.dispatchEvent(event)
	} finally {
		delete (event as any)[collector]
	}
	const settled = await Promise.allSettled(promises)
	for (const s of settled) {
		if (s.status === "rejected") {
			throw s.reason instanceof HookErr ? s.reason : new HookErr(hookErrKinds.other, s.reason)
		}
	}
	return settled.length
}
