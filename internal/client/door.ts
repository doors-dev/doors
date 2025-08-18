import { rootId } from './params'
import { doorId } from './lib'

import { attach as attachCaptures } from "./capture"
import navigator from "./navigator"
import { attach as attachDyna } from "./dyna"

type Handler = [(arg: any) => any, ((response: Response) => (void | Promise<void>)) | undefined]
type Closure = () => void | Promise<void>


const attr = "data-d00r"
const tag = "d0-0r"

type DoorElement = Element & {
    _d00r: {
        id: number
        parent: number
        impostor: boolean
    }
}

function execute(c: Closure) {

    const logError = (e: any) => {
        console.error("unmount handler error", e)
    }
    try {
        const result = c()
        if (result instanceof Promise) {
            result.then().catch(e => logError(e))
        }
    } catch (e) {
        logError(e)
    }
}
function getSelfId(el: Element): (number | undefined) {
    if (!el.matches(`${tag}, [${attr}]`)) {
        return undefined
    }
    return doorId(el.id)
}
function getParentId(el: Element): number {
    const parent = el.parentElement!.closest(`${tag}, [${attr}]`)
    if (!parent) {
        return rootId
    }
    return doorId(parent.id)
}

class Doors {
    private elements = new Map<number, DoorElement>()
    private handlers = new Map<number, Map<string, Handler>>()
    private onClear = new Map<number, Array<Closure>>()
    private onRemove = new Map<number, Array<Closure>>()
    private impostors = new Map<number, Set<number>>()

    private scanImpostors(parent: Element | Document) {
        for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="indexed"])`)) {
            element.setAttribute(attr, "indexed")
            this.register(element, true)
        }
    }

    private clear(id: number) {
        this.handlers.delete(id)
        const closures = this.onClear.get(id)
        if (closures === undefined) {
            return
        }
        this.onClear.delete(id)
        closures.forEach(c => execute(c))
        const impostors = this.impostors.get(id)
        if (impostors === undefined) {
            return
        }
        for (const impostor of impostors) {
            const element = this.elements.get(impostor)!
            this.unregister(element)
        }
    }

    register(element: Element, impostor: boolean = false) {
        const door = element as DoorElement
        const id = doorId(element.id);
        door._d00r = {
            id: id,
            parent: getParentId(element),
            impostor: impostor,
        }
        this.elements.set(id, door)
        if (!impostor) {
            return
        }
        let siblings = this.impostors.get(door._d00r.parent)
        if (!siblings) {
            siblings = new Set()
            this.impostors.set(door._d00r.parent, siblings)
        }
        siblings.add(id)
    }

    unregister(element: Element): void {
        const door = element as DoorElement
        this.elements.delete(door._d00r.id)
        this.clear(door._d00r.id)
        const onRemove = this.onRemove.get(door._d00r.id)
        if (onRemove !== undefined) {
            this.onRemove.delete(door._d00r.id)
            onRemove.forEach(c => execute(c))
        }
        if (!door._d00r.impostor) {
            return
        }
        const siblings = this.impostors.get(door._d00r.parent)!
        siblings.delete(door._d00r.id)
        if (siblings.size == 0) {
            this.impostors.delete(door._d00r.parent)
        }
    }

    scan(parent: Element | Document) {
        this.scanImpostors(parent)
        attachCaptures(parent)
        attachDyna(parent)
        navigator.activate(parent)
    }

    update(id: number, content: string) {
        const door = this.elements.get(id)
        if (!door) {
            throw new Error(`door ${id} not found`)
        }
        this.clear(door._d00r.id)

        const range = document.createRange()
        range.selectNodeContents(door)
        range.deleteContents()
        const fragment = range.createContextualFragment(content)
        navigator.activate(fragment)
        attachCaptures(fragment)

        range.insertNode(fragment)

        this.scanImpostors(door)
        attachDyna(door)
    }

    replace(id: number, content: string) {
        const door = this.elements.get(id)
        if (!door) {
            throw new Error(`door ${id} not found`)
        }
        if (door._d00r.impostor) {
            this.unregister(door)
        }
        const parent = door.parentElement!
        const range = document.createRange()
        range.selectNode(door)
        range.deleteContents()
        const fragment = range.createContextualFragment(content)

        navigator.activate(fragment)
        attachCaptures(fragment)

        range.insertNode(fragment)

        this.scanImpostors(parent)
        attachDyna(parent)
    }

    on(
        element: Element,
        name: string,
        handler: (arg: any) => any,
        response?: (response: Response) => void
    ): void {
        let id = getSelfId(element)
        if (id == undefined) {
            id = getParentId(element)
        }
        let handlers = this.handlers.get(id)
        if (!handlers) {
            handlers = new Map()
            this.handlers.set(id, handlers)
        }
        handlers.set(name, [handler, response])
    }

    onUnmount(element: Element, handler: () => void | Promise<void>): void {
        let id = getSelfId(element)
        if (id !== undefined) {
            if (!this.onRemove.has(id)) {
                this.onRemove.set(id, [handler])
                return
            }
            this.onRemove.get(id)!.push(handler)
            return
        }
        id = getParentId(element)
        if (!this.onClear.has(id)) {
            this.onClear.set(id, [handler])
            return
        }
        this.onClear.get(id)!.push(handler)
    }

    getHandler(id: number, name: string): Handler | undefined {
        const handlers = this.handlers.get(id)
        if (handlers !== undefined && handlers.has(name)) {
            return handlers.get(name)
        }
        if (id == rootId) {
            return undefined
        }
        const element = this.elements.get(id)
        if (element === undefined) {
            console.error("Unexpected behavior door [", id, "] not found for the call [" + name + "]")
            return undefined
        }
        return this.getHandler(element._d00r.parent, name)
    }

}

const doors = new Doors()

customElements.define(tag,
    class extends HTMLElement {
        constructor() {
            super()
        }
        connectedCallback() {
            doors.register(this)
        }
        disconnectedCallback() {
            doors.unregister(this)
        }
    }
)

export default doors
