// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

import doors from "./door"

class Dyna {
    private elements_ = new Set<Element>()
    constructor(private name: string) {

    }
    add(element: Element): boolean {
        if (this.elements_.has(element)) {
            return false
        }
        this.elements_.add(element)
        return true
    }
    remove(element: Element) {
        this.elements_.delete(element)
    }
    isEmpty(): boolean {
        return this.elements_.size === 0
    }
    removeAttr(): number {
        for (const element of this.elements_) {
            element.removeAttribute(this.name)
        }
        return this.elements_.size
    }
    setAttr(value: string): number {
        for (const element of this.elements_) {
            element.setAttribute(this.name, value)
        }
        return this.elements_.size
    }
}

class Registry {
    private registry_ = new Map<number, Dyna>()

    add(element: Element, id: number, name: string) {
        let dyna = this.registry_.get(id)
        if (!dyna) {
            dyna = new Dyna(name)
            this.registry_.set(id, dyna)
        }
        if (!dyna.add(element)) {
            return
        }
        doors.onUnmount(element, () => {
            dyna.remove(element)
            if (dyna.isEmpty()) {
                this.registry_.delete(id)
            }
        })
    }
    removeAttr(id: number): number {
        const dyna = this.registry_.get(id)
        if (!dyna) {
            return 0
        }
        return dyna.removeAttr()
    }
    setAttr(id: number, value: string): number {
        const dyna = this.registry_.get(id)
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

const attr = "data-d0y"
export function attach(parent: Element | DocumentFragment | Document) {
    for (const element of parent.querySelectorAll<Element>(`[${attr}]:not([${attr}="applied"])`)) {
        const dynaList = JSON.parse(element.getAttribute(attr)!)
        element.setAttribute(attr, "applied")
        for (const [id, name] of dynaList) {
            r.add(element, id, name)
        }
    }
}
