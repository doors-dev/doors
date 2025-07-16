import door from "./door"
import { capture, attach } from "./capture"
import navigator from "./navigator"
import { doorId } from "./lib"

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
    call: async ([name, arg, nodeId, hookId]: [string, any, number, number]) => {
        const entry = door.getHandler(nodeId, name)
        if (!entry) {
            throw new Error(`Handler ${name} not found`)
        }
        const [handler, response] = entry
        const result = await handler(arg)
        let requestResponse: any = undefined
        if (hookId !== null) {
            const hook = [nodeId, hookId, ["butter"], null, ""]
            requestResponse = await capture("default", result, null, hook)
        }
        if (response) {
            await response(requestResponse)
        }
    },

    set_path: ([p, replace]: [string, boolean]) => {
        if (replace) {
            navigator.replace(p)
            return
        }
        navigator.push(p)
    },

    node_replace: (nodeId: number, content: string) => {
        const node = document.getElementById(doorId(nodeId))
        if (!node) throw new Error("Node not found")

        const range = document.createRange()
        range.selectNode(node)
        range.deleteContents()

        const fragment = range.createContextualFragment(content)
        navigator.activateInside(fragment)
        attach(fragment)
        range.insertNode(fragment)
    },

    node_update: (nodeId: number, content: string) => {
        const node = document.getElementById(doorId(nodeId))
        if (!node) throw new Error("Node not found")

        const range = document.createRange()
        range.selectNodeContents(node)
        range.deleteContents()
        door.reset(nodeId)
        const fragment = range.createContextualFragment(content)
        navigator.activateInside(fragment)
        attach(fragment)
        range.insertNode(fragment)
    },

    node_remove: (nodeId: number) => {
        const node = document.getElementById(doorId(nodeId))
        if (!node) throw new Error("Node not found")
        node.remove()
    },

    touch: (_: any) => { }
}
