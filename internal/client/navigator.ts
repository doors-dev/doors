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

import { id, prefix, requestTimeout } from "./params"
import { AbortTimer, arraysEqual, scrollInto } from "./lib"
import indicator from "./indicator"
import doors from "./door";


type PathMatchType = "full" | "starts" | "parts";

const queryMatcherTypes = {
	all: "all",
	some: "some",
	ignoreAll: "ignore_all",
	ignoreSome: "ignore_some",
	ifPresent: "if"
} as const;

type QueryMatchType = typeof queryMatcherTypes[keyof typeof queryMatcherTypes]


export type LinkData = {
	url: URL,
	state: {
		indicator: number | undefined,
		url: URL | undefined,
	},
	settings: any
}

const linkState = Symbol()

export type LinkElement = Element & {
	[linkState]: LinkData
}


const urls = {
	toString(url: URL): string {
		let path = url.pathname
		if (!path.startsWith("/")) {
			path = "/" + path
		}
		return path + (url.search ? (path.endsWith("/") ? url.search : "/" + url.search) : "")
	},
	current(): URL {
		return new URL(window.location.href, window.location.origin)
	},
	new(path: string): URL {
		return new URL(path, window.location.origin)
	},
	equal(url1: URL, url2: URL, checkHash: boolean = false) {
		const hashMatch = checkHash ? url1.hash === url2.hash : true
		return hashMatch && url1.pathname === url2.pathname && this.searchEqual(this.searchToMap(url1.searchParams), this.searchToMap(url2.searchParams))
	},
	searchToMap(searchParams: URLSearchParams): Map<string, string[]> {
		const map = new Map<string, string[]>();
		for (const key of searchParams.keys()) {
			map.set(key, searchParams.getAll(key));
		}
		return map;
	},
	searchEqual(
		a: Map<string, string[]>,
		b: Map<string, string[]>
	): boolean {
		if (!arraysEqual([...a.keys()], [...b.keys()], false)) {
			return false;
		}
		for (const [key, value] of a.entries()) {
			if (!arraysEqual(value, b.get(key)!)) {
				return false;
			}
		}
		return true;
	}
}


const linkAttr = "data-d0a"

export class Navigator {
	private registry = new Set<LinkElement>()
	constructor(
	) {
		window.addEventListener('popstate', this.pop);
	}
	private blockPop: Array<{ id: number, time: number }> = []
	private isPopBlocked() {
		if (this.blockPop.length == 0) {
			return false
		}
		const blockPop = this.blockPop
		this.blockPop = []
		const id = history.state?._d0r?.next
		let blocked = false
		for (const block of blockPop) {
			if (Date.now() - block.time > 1000) {
				continue
			}
			if (id === block.id) {
				blocked = true
				continue
			}
			this.blockPop.push(block)
		}
		return blocked
	}
	private pop = async (): Promise<void> => {
		if (this.isPopBlocked()) {
			return
		}
		const abortTimer = new AbortTimer(requestTimeout)
		try {
			const r = await fetch(`${prefix}/u/${id}${urls.toString(urls.current())}`, {
				method: "GET",
				signal: abortTimer.signal,
			});
			if (!r.ok) {
				throw new Error("code " + r.status);
			}
		} catch (e) {
			location.reload()
		} finally {
			abortTimer.cancel()
		}
	}
	scan(parent: Element | Document | DocumentFragment) {
		const url = urls.current()
		for (const element of parent.querySelectorAll(`[${linkAttr}][href]:not([${linkAttr}="indexed"])`) as any as Array<HTMLAnchorElement>) {
			const settings = JSON.parse(element.getAttribute(linkAttr)!)
			element.setAttribute(linkAttr, "indexed")
			const linkElement = this.newLink(element, settings)
			this.activateLink(linkElement, url)
			this.registry.add(linkElement)
			doors.onUnmount(linkElement, () => {
				this.registry.delete(linkElement)
			})
		}
	}

	activate(url: URL) {
		for (const link of this.registry.values()) {
			this.activateLink(link, url)
		}
	}

	public activateCurrent(): void {
		this.activate(urls.current())
	}

	private trim(path: string): string {
		return path.replace(/^\/+|\/+$/g, "")
	}

	private newLink(el: Element, settings: any): LinkElement {
		const linkEl = el as LinkElement
		linkEl[linkState] = {
			url: urls.new(el.getAttribute("href")!),
			settings: settings,
			state: {
				indicator: undefined,
				url: undefined
			}
		}
		return linkEl
	}
	private activateLink(el: LinkElement, newUrl: URL): LinkElement {
		const data = el[linkState]
		const [pathMatchTuple, queryMatchers, matchFragment, indicators]: any = data.settings
		if (data.state.url && urls.equal(data.state.url, newUrl, matchFragment)) {
			return el
		}
		const pathMatch: PathMatchType = pathMatchTuple[0];
		const pathMatchArg: [number] | undefined = pathMatchTuple[1];
		const url = data.url
		let match = true;
		const newPath = this.trim(newUrl.pathname)
		const linkPath = this.trim(url.pathname)
		if (pathMatch === "full") {
			match = newPath === linkPath
		} else if (pathMatch === "starts") {
			match = newPath.startsWith(linkPath);
		} else if (pathMatch === "parts") {
			const newMatchPath = newPath.split("/")
			const linkMatchPath = linkPath.split("/")
			for (const index of pathMatchArg!) {
				if (newMatchPath[index] !== linkMatchPath[index]) {
					match = false
					break
				}
			}
		}
		if (match) {
			const newSearch = urls.searchToMap(newUrl.searchParams)
			const search = urls.searchToMap(url.searchParams)
			for (const matcher of queryMatchers) {
				const matchType: QueryMatchType = matcher[0];
				const matchArg: string[] | undefined = matcher[1];
				if (matchType == queryMatcherTypes.all) {
					match = urls.searchEqual(newSearch, search);
					break
				}
				if (matchType == queryMatcherTypes.ignoreAll) {
					break
				}
				if (matchType == queryMatcherTypes.ignoreSome) {
					for (const param of matchArg!) {
						newSearch.delete(param)
						search.delete(param)
					}
				}
				if (matchType == queryMatcherTypes.ifPresent || matchType == queryMatcherTypes.some) {
					const a = new Map()
					const b = new Map()
					for (const param of matchArg!) {
						if (matchType == queryMatcherTypes.ifPresent && !newSearch.has(param)) {
							search.delete(param)
							continue
						}
						a.set(param, newSearch.get(param))
						newSearch.delete(param)
						b.set(param, search.get(param))
						search.delete(param)
					}
					match = urls.searchEqual(a, b);
					if (!match) {
						break
					}
				}
			}
		}
		if (match && matchFragment) {
			match = url.hash == newUrl.hash
		}
		const prevIndicator = data.state.indicator
		let newIndicator: number | undefined = undefined
		if (match) {
			newIndicator = indicator.start(el, indicators)
		}
		data.state = { indicator: newIndicator, url: newUrl }
		indicator.end(prevIndicator)
		return el
	}

	private counter = 0
	public push(path: string, serverPush: boolean): (() => void) | null {
		const newUrl = urls.new(path)
		const currentUrl = urls.current()
		if (!urls.equal(currentUrl, newUrl)) {
			this.counter += 1
			const id = this.counter
			history.replaceState({ ...history.state, _d0r: { ...history.state?._d0r, next: id } }, '')
			history.pushState({ _d0r: { id } }, '', path);
			if (serverPush) {
				this.activate(newUrl)
			}
			return () => {
				if (history.state?._d0r?.id !== id) {
					return;
				}
				this.blockPop.push({ id, time: Date.now() })
				history.go(-1);
			}
		}
		if (serverPush) {
			history.replaceState({ ...history.state, _d0r: { ...history.state?._d0r, id: undefined } }, '')
			this.activateCurrent()
			return null
		}
		if (newUrl.hash == currentUrl.hash) {
			return null
		}
		history.replaceState({ ...history.state, _d0r: undefined }, '', path);
		this.activateCurrent()
		const hash = newUrl.hash != "" && newUrl.hash != "#" ? newUrl.hash : undefined
		if (hash) {
			scrollInto(hash)
		}
		return null
	}
	public replace(path: string, serverPush: boolean): (() => void) | null {
		const newUrl = urls.new(path)
		const currentUrl = urls.current()
		if (!urls.equal(currentUrl, newUrl)) {
			if (serverPush) {
				history.replaceState({ ...history.state, _d0r: undefined }, '', path)
				this.activate(newUrl)
				return null
			}
			this.counter += 1
			const id = this.counter
			const priorHref = window.location.href
			const priorState = history.state
			history.replaceState({ ...priorState, _d0r: { ...priorState?._d0r, id } }, '', path)
			return () => {
				if (history.state?._d0r?.id !== id) {
					return
				}
				history.replaceState(priorState, '', priorHref)
			}
		}
		if (serverPush) {
			history.replaceState({ ...history.state, _d0r: { ...history.state?._d0r, id: undefined } }, '')
			this.activateCurrent()
			return null
		}
		if (newUrl.hash == currentUrl.hash) {
			return null
		}
		history.replaceState({ ...history.state, _d0r: undefined }, '', path)
		this.activateCurrent()
		const hash = newUrl.hash != "" && newUrl.hash != "#" ? newUrl.hash : undefined
		if (hash) {
			scrollInto(hash)
		}
		return null
	}

}

export default new Navigator()
