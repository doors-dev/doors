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
import controller from './controller'

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

    clean = (handler: () => void | Promise<void>): void => {
        door.onUnmount(this.anchor, handler)
    }

    get G(): any {
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
        $d: $D,
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
    return f($d, $on, $data, $hook, $fetch, $G, $ready, $clean, HookErr)
}
export default init
