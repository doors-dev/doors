// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import door from './door'
import { capture } from './capture'
import { HookErr, hookErrKinds } from './capture'
import controller from './controller'

function getHookParams(element: HTMLElement, name: string): any | undefined {
    const attrName = `data-d00r-hook:${name}`
    const value = element.getAttribute(attrName)
    if (value == null) return undefined
    return JSON.parse(value)
}


const global: { [key: string]: any } = {}

class $D {
    constructor(private anchor: HTMLElement) { }

    clean = (handler: () => void | Promise<void>): void => {
        door.onUnmount(this.anchor, handler)
    }

    get G() {
        return global
    }

    get ready(): Promise<undefined> {
        return controller.ready
    }

    on = (name: string, handler: (arg: any) => any): void => {
        door.on(this.anchor, name, handler)
    }

    fetchHook = async (name: string, arg: any): Promise<Response> => {
        const hook = getHookParams(this.anchor, name)
        if (hook === undefined) {
            throw new HookErr(hookErrKinds.capture, new Error("hook " + name + " not found"))
        }
        return await capture("default", undefined, arg, undefined, hook)
    }

    hook = async (name: string, arg: any): Promise<any> => {
        const res = await this.fetchHook(name, arg)
        return await res.json()
    }

    data = (name: string): any => {
        const attrName = `data-d00r-data:${name}`
        const value = this.anchor.getAttribute(attrName)
        if (value == null) return undefined
        return JSON.parse(value)
    }
}

function init(
    anchor: HTMLElement,
    f: (
        $on: $D['on'],
        $data: $D['data'],
        $hook: $D['hook'],
        $fetch: $D['fetchHook'],
        $G: $D['G'],
        $ready: $D['ready'],
        $clean: $D['clean'],
        HookErr: any,
    ) => Promise<void> | void
) {
    const $d = new $D(anchor)
    const $on = $d.on
    const $data = $d.data
    const $hook = $d.hook
    const $fetch = $d.fetchHook
    const $G = $d.G
    const $ready = $d.ready
    const $clean = $d.clean
    const HookErr = $d.HookErr
    return f($on, $data, $hook, $fetch, $G, $ready, $clean, HookErr)
}
export default init
