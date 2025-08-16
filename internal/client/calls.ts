import door from "./door"
import { capture, attach as attachCaptures } from "./capture"
import navigator from "./navigator"
import { attach as attachDyna, removeAttr, setAttr } from "./dyna"
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
            const hook = [nodeId, hookId, [], [], ""]
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

    node_replace: (nodeId: number, content: string) => {
        const node = document.getElementById(doorId(nodeId))
        if (!node) throw new Error("Node not found")

        const parent = node.parentElement
        const range = document.createRange()
        range.selectNode(node)
        range.deleteContents()

        const fragment = range.createContextualFragment(content)
        navigator.activateInside(fragment)
        attachCaptures(fragment)
        range.insertNode(fragment)
        attachDyna(parent!)
    },

    node_update: (nodeId: number, content: string) => {
        const id = doorId(nodeId)
        const node = document.getElementById(id)
        if (!node) throw new Error("Node not found")
        door.reset(nodeId)
        const range = document.createRange()
        range.selectNodeContents(node)
        range.deleteContents()
        const fragment = range.createContextualFragment(content)
        navigator.activateInside(fragment)
        attachCaptures(fragment)
        range.insertNode(fragment)
        attachDyna(document.getElementById(id)!)
    },

    node_remove: (nodeId: number) => {
        const node = document.getElementById(doorId(nodeId))
        if (!node) throw new Error("Node not found")
        node.remove()
    },

    touch: (_: any) => { }
}
