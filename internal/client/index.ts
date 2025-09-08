// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import door from './door'
import { capture } from './capture'
import { HookErr, hookErrKinds } from './capture'

function getHookParams(element: HTMLElement, name: string): any | undefined {
    const attrName = `data-d00r-hook:${name}`
    const value = element.getAttribute(attrName)
    if (value == null) return undefined
    return JSON.parse(value)
}


const global: { [key: string]: any } = {}

class $D {
    HookErr = HookErr

    constructor(private anchor: HTMLElement) { }

    clean(handler: () => void | Promise<void>): void {
        door.onUnmount(this.anchor, handler)
    }

    get G() {
        return global
    }

    on(name: string, handler: (arg: any) => any): void {
        door.on(this.anchor, name, handler)
    }

    async rawHook(name: string, arg: any): Promise<Response> {
        const hook = getHookParams(this.anchor, name)
        if (hook === undefined) {
            throw new HookErr(hookErrKinds.capture, new Error("hook " + name + " not found"))
        }
        return await capture("default", undefined, arg, undefined, hook)
    }

    async hook(name: string, arg: any): Promise<any> {
        const res = await this.rawHook(name, arg)
        return await res.json()
    }

    data(name: string): any {
        const attrName = `data-d00r-data:${name}`
        const value = this.anchor.getAttribute(attrName)
        if (value == null) return undefined
        return JSON.parse(value)
    }
}

function init(achor: HTMLElement): $D
function init(anchor: HTMLElement, f: ($d: $D) => (Promise<void> | void)): void
function init(anchor: HTMLElement, f?: ($d: $D) => (Promise<void> | void)): (void | $D) {
    const $d = new $D(anchor)
    if (!f) {
        return $d
    }
    f($d)
}
export default init
export { HookErr }
