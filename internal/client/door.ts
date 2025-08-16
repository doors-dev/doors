import { rootId } from './params'
import { nodeId } from './lib'

type Handler = [(arg: any) => any, ((response: Response) => (void | Promise<void>)) | undefined]



class Door {
    private doors: Doors
    private handlers: Map<string, Handler>
    private parent: number | null
    private unmountHandlers: (() => void | Promise<void>)[]
    constructor(
        doors: Doors,
        parent: number | null,
        handlers: Map<string, Handler> = new Map()
    ) {
        this.doors = doors
        this.handlers = handlers
        this.parent = parent
        this.unmountHandlers = []
    }
    on(
        name: string,
        handler: (arg: any) => any,
        response?: (response: Response) => void
    ): void {
        this.handlers.set(name, [handler, response])
    }
    onRemove(handler: () => void | Promise<void>): void {
        this.unmountHandlers.push(handler)
    }
    getHandler(name: string): Handler | undefined {
        const entry = this.handlers.get(name)
        if (entry) {
            const [handler, response] = entry
            return [handler, response]
        }
        if (this.parent === null) {
            return undefined
        }
        return this.doors.getHandler(this.parent, name)
    }
    clear() {
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
        this.handlers.clear()
        this.unmountHandlers = []
    }
}


class Doors {
    private doors: Map<number, Door>
    private bufferedHandlers: Map<number, Map<string, Handler>>
    private rootId: number

    constructor(rootId: number) {
        this.rootId = rootId
        this.doors = new Map()
        this.bufferedHandlers = new Map()
        this.doors.set(rootId, new Door(this, null))
    }

    register(element: Element): void {
        const id = nodeId(element.id)
        const handlers = this.bufferedHandlers.get(id)
        this.bufferedHandlers.delete(id)
        const parent = this.getParentId(element)
        const door = new Door(this, parent, handlers)
        this.doors.set(id, door)
    }

    unregister(element: Element): void {
        const id = nodeId(element.id)
        const door = this.doors.get(id)
        this.doors.delete(id)
        door!.clear()
    }

    private getParentId(el: Element): number {
        const parent = el.parentElement!.closest("do-or")
        if (!parent) {
            return this.rootId
        }
        return nodeId(parent.id)
    }
    on(
        script: Element,
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
            buffer.set(name, [handler, response])
            return
        }
        door.on(name, handler, response)
    }

    onRemove(element: Element, handler: () => void | Promise<void>): void {
        const id = this.getParentId(element)
        const door = this.doors.get(id)
        door!.onRemove(handler)
    }

    getHandler(id: number, name: string): Handler | undefined {
        const door = this.doors.get(id)
        return door?.getHandler(name)
    }
    reset(id: number): void {
        this.doors.get(id)?.clear()
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
            doors.unregister(this)
        }
    }
)

export default doors
