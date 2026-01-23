// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import doors from "./door"
import navigator from "./navigator"
import indicator, { IndicatorEntry } from "./indicator"
import { removeAttr, setAttr } from "./dyna"
import { HookErr } from "./capture"
import { doAfter } from "./lib"
import { report } from "./scope.ts"


type Options = {
    error?: HookErr,
    payload?: string,
    element?: Element,
}

const actions = {
    location_reload: (_: Options) => {
        doAfter(() => {
            location.reload()
        })
    },
    indicate: (opt: Options, duration: number, indicatations: Array<IndicatorEntry>) => {
        const id = indicator.start(opt.element ?? null, indicatations)
        if (id) {
            setTimeout(() => indicator.end(id), duration)
        }
    },
    report_hook: (_: Options, track: number) => {
        report(track)
    },
    location_replace: (_: Options, href: string, origin: boolean) => {
        let url: URL
        if (origin) {
            url = new URL(href, window.location.origin);
        } else {
            url = new URL(href)
        }
        doAfter(() => {
            location.replace(url.toString())
        })
    },
    scroll: (_: Options, selector: string, smooth: boolean) => {
        const el = document.querySelector(selector)
        if (el) {
            el.scrollIntoView({ behavior: smooth ? "smooth" : "auto" });
        }
    },
    location_assign: (_: Options, href: string, origin: boolean) => {
        let url: URL
        if (origin) {
            url = new URL(href, window.location.origin);
        } else {
            url = new URL(href)
        }
        doAfter(() => {
            location.assign(url.toString())
        })
    },
    emit: (opt: Options, name: string, arg: any, doorId: number): any => {
        const handler = doors.getHandler(doorId, name)
        if (!handler) {
            throw new Error(`Handler ${name} not found`)
        }
        return handler(arg, opt.error as any)
    },
    dyna_set: (_: Options, id: number, value: string) => {
        setAttr(id, value)
    },
    dyna_remove: (_: Options, id: number) => {
        removeAttr(id)
    },
    set_path: (_: Options, path: string, replace: boolean) => {
        if (replace) {
            navigator.replace(path)
            return
        }
        navigator.push(path)
    },
    door_replace: (opt: Options, doorId: number) => {
        doors.replace(doorId, opt.payload!)
    },
    door_update: (opt: Options, doorId: number) => {
        doors.update(doorId, opt.payload!)
    },
}

type Output = Exclude<any, undefined>;
type Err = {
    message: string;
    [key: string]: any;
};

export type CallResult = ([Output, undefined] | [undefined, Err])

export type Action = [string, Array<any>]

export default function action(name: string, args: Array<any>, options: Options = {}): CallResult {
    try {
        const fn = actions[name]
        if (!fn) {
            throw new Error(`action [${name}] not found`)
        }
        let output = fn(options, ...args)
        if (output instanceof Promise) {
            throw new Error("async actions are prohibited")
        }
        if (output === undefined) {
            output = null
        }
        return [output, undefined]
    } catch (e) {
        if (e && typeof e === "object" && typeof e.message === "string") {
            return [undefined, e]
        }
        return [undefined, new Error("unknown error")]
    }
}
