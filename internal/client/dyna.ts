// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import doors from "./door"

class Dyna {
    private elements = new Set<Element>()
    constructor(private name: string) {

    }
    add(element: Element): boolean {
        if (this.elements.has(element)) {
            return false
        }
        this.elements.add(element)
        return true
    }
    remove(element: Element) {
        this.elements.delete(element)
    }
    isEmpty(): boolean {
        return this.elements.size === 0
    }
    removeAttr(): number {
        for (const element of this.elements) {
            element.removeAttribute(this.name)
        }
        return this.elements.size
    }
    setAttr(value: string): number {
        for (const element of this.elements) {
            element.setAttribute(this.name, value)
        }
        return this.elements.size
    }
}

class Registry {
    private registry = new Map<number, Dyna>()

    add(element: Element, id: number, name: string) {
        let dyna = this.registry.get(id)
        if (!dyna) {
            dyna = new Dyna(name)
            this.registry.set(id, dyna)
        }
        if (!dyna.add(element)) {
            return
        }
        doors.onUnmount(element, () => {
            dyna.remove(element)
            if (dyna.isEmpty()) {
                this.registry.delete(id)
            }
        })
    }
    removeAttr(id: number): number {
        const dyna = this.registry.get(id)
        if (!dyna) {
            return 0
        }
        return dyna.removeAttr()
    }
    setAttr(id: number, value: string): number {
        const dyna = this.registry.get(id)
        if (!dyna) {
            return 0
        }
        return dyna.setAttr(value)
    }
}


const r = new Registry()


export function removeAttr(id: number): number {
    return r.removeAttr(id)
}
export function setAttr(id: number, value: string): number {
    return r.setAttr(id, value)
}

const attr = "data-d00r-dyna"
export function attach(parent: Element | DocumentFragment | Document) {
    for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="applied"])`)) {
        const dynaList = JSON.parse(element.getAttribute(attr)!)
        element.setAttribute(attr, "applied")
        for (const [id, name] of dynaList) {
            r.add(element, id, name)
        }
    }
}
