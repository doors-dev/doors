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

import { id, prefix } from "./params"
import { arraysEqual, scrollInto } from "./lib"
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

export type LinkElement = Element & {
	_d0r: LinkData
}


const linkAttr = "data-d0a"

export class Navigator {
	private registry = new Set<LinkElement>()

	constructor(
	) {
		window.addEventListener('popstate', async () => {
			await this.pop();
		});
	}

	private searchToMap(searchParams: URLSearchParams): Map<string, string[]> {
		const map = new Map<string, string[]>();
		for (const key of searchParams.keys()) {
			map.set(key, searchParams.getAll(key));
		}
		return map;
	}

	private searchEqual(
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

	scan(parent: Element | Document | DocumentFragment) {
		const url = this.urlCurrent()
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
		this.activate(this.urlCurrent())
	}

	private trim(path: string): string {
		return path.replace(/^\/+|\/+$/g, "")
	}

	private newLink(el: Element, settings: any): LinkElement {
		const linkEl = el as LinkElement
		linkEl._d0r = {
			url: new URL(el.getAttribute("href")!, window.location.origin),
			settings: settings,
			state: {
				indicator: undefined,
				url: undefined
			}
		}
		return linkEl
	}
	private activateLink(el: LinkElement, newUrl: URL): LinkElement {
		const data = el._d0r
		const [pathMatchTuple, queryMatchers, matchFragment, indicators]: any = data.settings
		if (data.state.url && this.urlAreEqual(data.state.url, newUrl, matchFragment)) {
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
			const newSearch = this.searchToMap(newUrl.searchParams)
			const search = this.searchToMap(url.searchParams)
			for (const matcher of queryMatchers) {
				const matchType: QueryMatchType = matcher[0];
				const matchArg: string[] | undefined = matcher[1];
				if (matchType == queryMatcherTypes.all) {
					match = this.searchEqual(newSearch, search);
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
					match = this.searchEqual(a, b);
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

	private async pop(): Promise<void> {
		try {
			const r = await fetch(`${prefix}/u/${id}${this.urlCurrentString()}`, {
				method: "GET",
			});
			if (!r.ok) {
				throw new Error("code " + r.status);
			}
		} catch (e) {
			location.reload()
		}
	}

	urlCurrent(): URL {
		return new URL(window.location.href)
	}

	private urlCurrentString(): string {
		const url = this.urlCurrent()
		let path = url.pathname
		if (!path.startsWith("/")) {
			path = "/" + path
		}
		return path + (url.search ? (path.endsWith("/") ? url.search : "/" + url.search) : "")
	}

	private urlAreEqual(url1: URL, url2: URL, checkHash: boolean = false) {
		const hashMatch = checkHash ? url1.hash === url2.hash : true
		return hashMatch && url1.pathname === url2.pathname && this.searchEqual(this.searchToMap(url1.searchParams), this.searchToMap(url2.searchParams))
	}
	public push(path: string, serverPush: boolean): boolean {
		const newUrl = new URL(path, window.location.origin);
		const currentUrl = new URL(this.urlCurrent(), window.location.origin)
		if (!this.urlAreEqual(currentUrl, newUrl)) {
			history.pushState(null, '', path);
			if (serverPush) {
				this.activateCurrent()
			}
			return true
		}
		if (serverPush) {
			this.activateCurrent()
			return false
		}
		if (newUrl.hash != currentUrl.hash) {
			history.replaceState(null, '', path);
			this.activateCurrent()
		}
		const hash = newUrl.hash != "" && newUrl.hash != "#" ? newUrl.hash : undefined
		if (hash) {
			scrollInto(hash)
		}
		return false
	}
	public replace(path: string): void {
		const newUrl = new URL(path, window.location.origin);
		const currentUrl = new URL(this.urlCurrent(), window.location.origin)
		if (!this.urlAreEqual(currentUrl, newUrl)) {
			this.activate(newUrl)
			history.replaceState(null, '', path);
		}
	}

}

export default new Navigator()
