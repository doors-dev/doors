import { rootId } from './params'
import { nodeId } from './lib'

type Handler = [(arg: any) => any, ((response: Response) => (void | Promise<void>)) | undefined]

type HandlerEntry = [HTMLElement, ...Handler]


class Door {
    private element: HTMLElement
    private doors: Doors
    private handlers: Map<string, HandlerEntry>
    private parent: number | null
    private unmountHandlers: (() => void | Promise<void>)[]

    constructor(
        doors: Doors,
        parent: number | null,
        element: HTMLElement,
        handlers: Map<string, HandlerEntry> = new Map()
    ) {
        this.element = element
        this.doors = doors
        this.handlers = handlers
        this.parent = parent
        this.unmountHandlers = []
    }
    same(element: HTMLElement): boolean {
        return element === this.element
    }
    on(
        script: HTMLElement,
        name: string,
        handler: (arg: any) => any,
        response?: (response: Response) => void
    ): void {
        this.handlers.set(name, [script, handler, response])
    }
    onRemove(handler: () => void | Promise<void>): void {
        this.unmountHandlers.push(handler)
    }
    getHandler(name: string): Handler | undefined {
        const entry = this.handlers.get(name)
        if (entry) {
            const [, handler, response] = entry
            return [handler, response]
        }
        if (this.parent === null) {
            return undefined
        }
        return this.doors.getHandler(this.parent, name)
    }

    reset(): void {
        this.handlers.clear()
        this.unmountHandlers = []
    }
    clear(): Map<string, HandlerEntry> {
        if (this.element.isConnected) {
            this.element.remove()
            const logError = (e: any) => {
                console.error("unmount handler error", e)
            }
            for (const handler of this.unmountHandlers) {
                try {
                    const result = handler()
                    if (result instanceof Promise) {
                        result.then().catch(e => logError(e))
                    }
                } catch (e) {
                    logError(e)
                }
            }
        }

        for (const [key, [script]] of this.handlers) {
            if (!script.isConnected) {
                this.handlers.delete(key)
            }
        }
        return this.handlers
    }
}


class Doors {
    private doors: Map<number, Door>
    private bufferedHandlers: Map<number, Map<string, HandlerEntry>>
    private rootId: number

    constructor(rootId: number) {
        this.rootId = rootId
        this.doors = new Map()
        this.bufferedHandlers = new Map()
        this.doors.set(rootId, new Door(this, null, document.documentElement))
    }

    register(element: HTMLElement): void {
        const id = nodeId(element.id)
        const existing = this.doors.get(id)
        let handlers: Map<string, HandlerEntry> | undefined

        if (existing) {
            handlers = existing.clear()
        } else {
            handlers = this.bufferedHandlers.get(id)
            this.bufferedHandlers.delete(id)
        }

        const parent = this.getParentId(element)
        const door = new Door(this, parent, element, handlers)
        this.doors.set(id, door)
    }

    reset(id: number): void {
        this.doors.get(id)?.reset()
    }

    remove(element: HTMLElement): void {
        const id = nodeId(element.id)
        const door = this.doors.get(id)
        if (door?.same(element)) {
            door.clear()
            this.doors.delete(id)
        }
    }

    private getParentId(el: HTMLElement): number {
        const parent = el.parentElement!.closest("do-or")
        if (!parent) {
            return this.rootId
        }
        return nodeId(parent.id)
    }
    on(
        script: HTMLElement,
        name: string,
        handler: (arg: any) => any,
        response?: (response: Response) => void
    ): void {
        const id = this.getParentId(script)
        const door = this.doors.get(id)
        if (!door) {
            let buffer = this.bufferedHandlers.get(id)
            if (!buffer) {
                buffer = new Map()
                this.bufferedHandlers.set(id, buffer)
            }
            buffer.set(name, [script, handler, response])
            return
        }
        door.on(script, name, handler, response)
    }

    onRemove(script: HTMLElement, handler: () => void | Promise<void>): void {
        const id = this.getParentId(script)
        const door = this.doors.get(id)
        door?.onRemove(handler)
    }

    getHandler(id: number, name: string): Handler | undefined {
        const door = this.doors.get(id)
        return door?.getHandler(name)
    }
}

const doors = new Doors(rootId)

customElements.define('do-or',
    class extends HTMLElement {
        constructor() {
            super()
        }
        connectedCallback() {
            doors.register(this)
        }
        disconnectedCallback() {
            doors.remove(this)
        }
    }
)

export default doors
