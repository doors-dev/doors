// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { setDiff, arrayDiff, splitClass } from "./lib"


type SelectorType = "target" | "query" | "query_all" | "parent_query"
type Kind = "attr" | "class" | "remove_class" | "content"


export type IndicatorEntry = [[SelectorType, string | null], Kind, string, string | null | undefined]

interface Indication {
    attrs: Map<string, string | null>
    classes: Set<string>
    removeClasses: Set<string>
    content: string | null
}

function newIndicator(
    target: Element | null,
    indicators: IndicatorEntry[]
): Map<Element, Indication> | undefined {
    const indications = new Map<Element, Indication>()
    for (const entry of indicators) {
        const [[selectorType, query], kind, param1, param2] = entry
        const elements: Array<Element> = []
        if (selectorType === "query") {
            const element = document.querySelector(query!)
            if (element) {
                elements.push(element)
            }
        } else if (selectorType === "query_all") {
            elements.push(...document.querySelectorAll(query!))
        } else if (selectorType === "parent_query") {
            if (target && target.parentElement) {
                const anchestor = target.parentElement.closest(query!)
                if (anchestor) {
                    elements.push(anchestor)
                }
            }
        } else if (target) {
            elements.push(target)
        }
        for (const el of elements) {
            let indication = indications.get(el)
            if (!indication) {
                indication = {
                    attrs: new Map<string, string | null>(),
                    classes: new Set(),
                    removeClasses: new Set(),
                    content: null,
                }
                indications.set(el, indication)
            }
            switch (kind) {
                case "attr":
                    indication.attrs.set(param1, param2 ?? null)
                    break
                case "class":
                    splitClass(param1).forEach(c => indication.classes.add(c))
                    break
                case "remove_class":
                    splitClass(param1).forEach(c => indication.removeClasses.add(c))
                    break
                case "content":
                    indication.content = param1
                    break
            }
        }
    }

    return indications.size === 0 ? undefined : indications
}

class ElementIndicator {
    private saved = {
        content: null as string | null,
        attrs: new Map<string, string | null>(),
    }

    private active: [number, Indication] | null = null
    private queue: [number, Indication][] = []

    constructor(private el: Element) { }

    private applyNext(
        id: number,
        indication: Indication,
        classToRemove: string[],
        classToAdd: string[],
        attrsToReset: string[],
        attrsToSet: string[]
    ) {
        for (const remove of classToRemove) {
            this.el.classList.remove(remove)
        }
        for (const add of classToAdd) {
            this.el.classList.add(add)
        }
        for (const reset of attrsToReset) {
            this.restoreAttr(reset)
        }
        for (const set of attrsToSet) {
            if (!this.saved.attrs.has(set)) {
                this.saved.attrs.set(set, this.el.getAttribute(set))
            }
            const value = indication.attrs.get(set)
            if (value === null) {
                this.el.removeAttribute(set)
            } else {
                this.el.setAttribute(set, value!)
            }
        }
        this.active = [id, indication]
    }

    private restoreAttr(name: string) {
        const saved = this.saved.attrs.get(name)
        this.saved.attrs.delete(name)
        if (saved === undefined || saved === null) {
            this.el.removeAttribute(name)
        } else {
            this.el.setAttribute(name, saved)
        }
    }

    start(id: number, indication: Indication) {
        if (this.active != null) {
            this.queue.push([id, indication])
            return
        }
        if (indication.content != null) {
            this.saved.content = this.el.innerHTML
            this.el.innerHTML = indication.content
        }
        this.applyNext(id, indication, [...indication.removeClasses], [...indication.classes], [], [...indication.attrs.keys()])
    }

    end(id: number): boolean {
        const [activeId, activeIndication] = this.active!
        if (id !== activeId) {
            this.queue = this.queue.filter(([queueId]) => id !== queueId)
            return false
        }

        const next = this.queue.shift()
        if (!next) {
            if (activeIndication.content !== null) {
                this.el.innerHTML = this.saved.content ?? ""
            }

            for (const name of activeIndication.attrs.keys()) {
                this.restoreAttr(name)
            }

            for (const remove of activeIndication.classes) {
                this.el.classList.remove(remove)
            }
            for (const remove of activeIndication.removeClasses) {
                this.el.classList.add(remove)
            }

            return true
        }

        const [nextId, nextIndication] = next

        const [classToRemove1, classToAdd1] = setDiff(
            activeIndication.classes,
            nextIndication.classes
        )

        const [classToAdd2, classToRemove2] = setDiff(
            activeIndication.removeClasses,
            nextIndication.removeClasses
        )

        const [attrsToReset] = arrayDiff(
            [...activeIndication.attrs.keys()],
            [...nextIndication.attrs.keys()]
        )

        const attrsToSet = [...nextIndication.attrs.entries()]
            .filter(([key, value]) => value !== activeIndication.attrs.get(key))
            .map(([key]) => key)

        if (activeIndication.content !== null) {
            if (nextIndication.content === null) {
                this.el.innerHTML = this.saved.content ?? ""
                this.saved.content = null
            } else if (nextIndication.content !== activeIndication.content) {
                this.el.innerHTML = nextIndication.content
            }
        } else if (nextIndication.content !== null) {
            this.saved.content = this.el.innerHTML
            this.el.innerHTML = nextIndication.content
        }

        this.applyNext(nextId, nextIndication, [...classToRemove1, ...classToRemove2], [...classToAdd1, ...classToAdd2], attrsToReset, attrsToSet)
        return false
    }
}

class IndicationController {
    private indicators = new Map<number, Map<Element, Indication>>()
    private elements = new WeakMap<Element, ElementIndicator>()
    private counter = 0

    start(target: Element | null, indicators: IndicatorEntry[] | null): number | undefined {
        if (!indicators || indicators.length === 0) {
            return undefined
        }
        const indicator = newIndicator(target, indicators)
        if (!indicator) {
            return undefined
        }
        this.counter += 1
        this.indicators.set(this.counter, indicator)
        for (const [el, indication] of indicator.entries()) {
            let element = this.elements.get(el)
            if (!element) {
                element = new ElementIndicator(el)
                this.elements.set(el, element)
            }
            element.start(this.counter, indication)
        }
        return this.counter
    }

    end(id: number | undefined): void {
        if (id === undefined) return
        const indication = this.indicators.get(id)
        if (!indication) return
        this.indicators.delete(id)
        for (const el of indication.keys()) {
            if (!el.isConnected) {
                this.elements.delete(el)
                continue
            }
            const element = this.elements.get(el)
            if (!element) continue
            const done = element.end(id)
            if (done) {
                this.elements.delete(el)
            }
        }
    }
}


export default new IndicationController()
