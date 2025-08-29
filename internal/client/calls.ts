// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import doors from "./door"
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
    call: ([name, arg, doorId]: [string, any, number]): any => {
        const handler = doors.getHandler(doorId, name)
        if (!handler) {
            throw new Error(`Handler ${name} not found`)
        }
        return handler(arg)
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
