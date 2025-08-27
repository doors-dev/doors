// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import doors from "./door"
import { capture } from "./capture"
import navigator from "./navigator"
import { removeAttr, setAttr } from "./dyna"

const task = (f: () => void) => {
    setTimeout(f, 0)
}

export default {
    location_reload: () => {
        task(() => {
            location.reload()
        })
    },
    location_replace: ([href, origin]: [string, boolean]) => {
        let url: URL
        if (origin) {
            url = new URL(href, window.location.origin);
        } else {
            url = new URL(href)
        }
        task(() => {
            location.replace(url.toString())
        })
    },
    scroll_into: ([selector, smooth]: [string, boolean]) => {
        const el = document.querySelector(selector)
        if (el) {
            el.scrollIntoView({ behavior: smooth ? "smooth" : "auto" });
        }
    },
    location_assign: ([href, origin]: [string, boolean]) => {
        let url: URL
        if (origin) {
            url = new URL(href, window.location.origin);
        } else {
            url = new URL(href)
        }
        task(() => {
            location.assign(url.toString())
        })
    },
    call: async ([name, arg, doorId, hookId]: [string, any, number, number]) => {
        const entry = doors.getHandler(doorId, name)
        if (!entry) {
            throw new Error(`Handler ${name} not found`)
        }
        const [handler, response] = entry
        const result = await handler(arg)
        let requestResponse: any = undefined
        if (hookId !== null) {
            const hook = [doorId, hookId, [], [], ""]
            requestResponse = await capture("default", null, result, undefined, hook)
        }
        if (response) {
            await response(requestResponse)
        }
    },
    dyna_set: ([id, value]: [number, string]) => {
        setAttr(id, value)
    },
    dyna_remove: (id: number) => {
        removeAttr(id)
    },
    set_path: ([p, replace]: [string, boolean]) => {
        if (replace) {
            navigator.replace(p)
            return
        }
        navigator.push(p)
    },

    door_replace: (doorId: number, content: string) => {
        doors.replace(doorId, content)
    },

    door_update: (doorId: number, content: string) => {
        doors.update(doorId, content)
    },


    touch: (_: any) => { }
}
