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

const trigger = Symbol()

type TriggerMeta = {
	hook: number
	consumed: boolean
	promise?: Promise<Response>
}

export function stampTrigger(event: Event, hook: number) {
	(event as any)[trigger] = { hook, consumed: false }
}

export function peekTrigger(event: Event): TriggerMeta | undefined {
	return (event as any)[trigger]
}

export function consumeTrigger(event: Event): TriggerMeta | undefined {
	const value: TriggerMeta | undefined = (event as any)[trigger]
	if (value === undefined) {
		return undefined
	}
	value.consumed = true
	return value
}

export function clearTrigger(event: Event): TriggerMeta | undefined {
	const value = (event as any)[trigger]
	delete (event as any)[trigger]
	return value
}
