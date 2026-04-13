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


export type IndicatorEntry = [[SelectorType, string | null], Kind, string, string | undefined]

interface Indication {
    attrs_: Map<string, string>
    classes_: Set<string>
    removeClasses_: Set<string>
    content_: string | null
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
                    attrs_: new Map<string, string>(),
                    classes_: new Set(),
                    removeClasses_: new Set(),
                    content_: null,
                }
                indications.set(el, indication)
            }
            switch (kind) {
                case "attr":
                    indication.attrs_.set(param1, param2!)
                    break
                case "class":
                    splitClass(param1).forEach(c => indication.classes_.add(c))
                    break
                case "remove_class":
                    splitClass(param1).forEach(c => indication.removeClasses_.add(c))
                    break
                case "content":
                    indication.content_ = param1
                    break
            }
        }
    }

    return indications.size === 0 ? undefined : indications
}

class ElementIndicator {
    private saved_ = {
        content_: null as string | null,
        attrs_: new Map<string, string>(),
    }

    private active_: [number, Indication] | null = null
    private queue_: [number, Indication][] = []

    constructor(private el_: Element) { }

    private applyNext(
        id: number,
        indication: Indication,
        classToRemove: string[],
        classToAdd: string[],
        attrsToReset: string[],
        attrsToSet: string[]
    ) {
        for (const remove of classToRemove) {
            this.el_.classList.remove(remove)
        }
        for (const add of classToAdd) {
            this.el_.classList.add(add)
        }
        for (const reset of attrsToReset) {
            const saved = this.saved_.attrs_.get(reset)
            if (saved === undefined) {
                this.el_.removeAttribute(reset)
            } else {
                this.saved_.attrs_.delete(reset)
                this.el_.setAttribute(reset, saved)
            }
        }
        for (const set of attrsToSet) {
            if (!this.saved_.attrs_.has(set)) {
                const toSave = this.el_.getAttribute(set)
                if (toSave !== null) {
                    this.saved_.attrs_.set(set, toSave)
                }
            }
            this.el_.setAttribute(set, indication.attrs_.get(set)!)
        }
        this.active_ = [id, indication]
    }

    start(id: number, indication: Indication) {
        if (this.active_ != null) {
            this.queue_.push([id, indication])
            return
        }
        if (indication.content_ != null) {
            this.saved_.content_ = this.el_.innerHTML
            this.el_.innerHTML = indication.content_
        }
        this.applyNext(id, indication, [...indication.removeClasses_], [...indication.classes_], [], [...indication.attrs_.keys()])
    }

    end(id: number): boolean {
        const [activeId, activeIndication] = this.active_!
        if (id !== activeId) {
            this.queue_ = this.queue_.filter(([queueId]) => id !== queueId)
            return false
        }

        const next = this.queue_.shift()
        if (!next) {
            if (activeIndication.content_ !== null) {
                this.el_.innerHTML = this.saved_.content_ ?? ""
            }

            if (activeIndication.attrs_.size !== 0) {
                for (const name of activeIndication.attrs_.keys()) {
                    const saved = this.saved_.attrs_.get(name)
                    if (saved === undefined) {
                        this.el_.removeAttribute(name)
                    } else {
                        this.el_.setAttribute(name, saved)
                    }
                }
            }

            for (const remove of activeIndication.classes_) {
                this.el_.classList.remove(remove)
            }
            for (const remove of activeIndication.removeClasses_) {
                this.el_.classList.add(remove)
            }

            return true
        }

        const [nextId, nextIndication] = next

        const [classToRemove1, classToAdd1] = setDiff(
            activeIndication.classes_,
            nextIndication.classes_
        )

        const [classToAdd2, classToRemove2] = setDiff(
            activeIndication.removeClasses_,
            nextIndication.removeClasses_
        )

        const [attrsToReset] = arrayDiff(
            [...activeIndication.attrs_.keys()],
            [...nextIndication.attrs_.keys()]
        )

        const attrsToSet = [...nextIndication.attrs_.entries()]
            .filter(([key, value]) => value !== activeIndication.attrs_.get(key))
            .map(([key]) => key)

        if (activeIndication.content_ !== null) {
            if (nextIndication.content_ === null) {
                this.el_.innerHTML = this.saved_.content_ ?? ""
                this.saved_.content_ = null
            } else if (nextIndication.content_ !== activeIndication.content_) {
                this.el_.innerHTML = nextIndication.content_
            }
        } else if (nextIndication.content_ !== null) {
            this.saved_.content_ = this.el_.innerHTML
            this.el_.innerHTML = nextIndication.content_
        }

        this.applyNext(nextId, nextIndication, [...classToRemove1, ...classToRemove2], [...classToAdd1, ...classToAdd2], attrsToReset, attrsToSet)
        return false
    }
}

class IndicationController {
    private indicators_ = new Map<number, Map<Element, Indication>>()
    private elements_ = new WeakMap<Element, ElementIndicator>()
    private counter_ = 0

    start(target: Element | null, indicators: IndicatorEntry[] | null): number | undefined {
        if (!indicators || indicators.length === 0) {
            return undefined
        }
        const indicator = newIndicator(target, indicators)
        if (!indicator) {
            return undefined
        }
        this.counter_ += 1
        this.indicators_.set(this.counter_, indicator)
        for (const [el, indication] of indicator.entries()) {
            let element = this.elements_.get(el)
            if (!element) {
                element = new ElementIndicator(el)
                this.elements_.set(el, element)
            }
            element.start(this.counter_, indication)
        }
        return this.counter_
    }

    end(id: number | undefined): void {
        if (id === undefined) return
        const indication = this.indicators_.get(id)
        if (!indication) return
        this.indicators_.delete(id)
        for (const el of indication.keys()) {
            if (!el.isConnected) {
                this.elements_.delete(el)
                continue
            }
            const element = this.elements_.get(el)
            if (!element) continue
            const done = element.end(id)
            if (done) {
                this.elements_.delete(el)
            }
        }
    }
}


export default new IndicationController()
