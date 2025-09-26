// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

import { detached, id } from "./params"
import { arraysEqual} from "./lib"
import indicator from "./indicator"


type PathMatchType = "full" | "starts" | "parts";

const queryMatcherTypes = {
    all: "all",
    some: "some",
    ignoreAll: "ignore_all",
    ignoreSome: "ignore_some",
    ifPresent: "if"
} as const;

type QueryMatchType = typeof queryMatcherTypes[keyof typeof queryMatcherTypes]


type Cache = {
    indicator: number | undefined,
    url: URL,
    settings: any
}
const attrName = "data-d00r-active"

export class Navigator {
    private cache = new WeakMap<Element, Cache>()
    constructor(
        private id: string,
    ) {
        window.addEventListener('popstate', async () => {
            await this.update(this.urlCurrent());
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


    public activate(e: Element | DocumentFragment | Document): void {
        this.activateLinks(this.urlCurrent(), e)
    }

    private trim(path: string): string {
        return path.replace(/^\/+|\/+$/g, "")
    }

    private activateLinks(newUrl: URL, parent: Element | DocumentFragment | Document = document): void {
        const links = parent.querySelectorAll(`[${attrName}]`) as any as Array<HTMLAnchorElement>;
        links.forEach(link => {
            const href = link.getAttribute("href")
            if (href === null) {
                return
            }
            let cache = this.cache.get(link)
            if (!!cache && this.urlAreEqual(cache.url, newUrl)) {
                return
            }
            let settings = cache?.settings
            if (!settings) {
                settings = JSON.parse(link.getAttribute(attrName)!)
                link.setAttribute(attrName, "cached")
            }
            const [pathMatchTuple, queryMatchers, indicators]: any = settings;
            const pathMatch: PathMatchType = pathMatchTuple[0];
            const pathMatchArg: number | undefined = pathMatchTuple[1];
            const url = new URL(href, window.location.origin);
            let match = false;
            const newPath = this.trim(newUrl.pathname)
            const linkPath = this.trim(url.pathname)
            if (pathMatch === "full") {
                match = newPath === linkPath
            } else if (pathMatch === "starts") {
                match = newPath.startsWith(linkPath);
            } else if (pathMatch === "parts" && pathMatchArg !== undefined) {
                const newMatchPath = newPath.split("/").slice(0, pathMatchArg).join("/");
                const linkMatchPath = linkPath.split("/").slice(0, pathMatchArg).join("/");
                match = newMatchPath === linkMatchPath;
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
            const prevIndicator = cache?.indicator
            if (match) {
                this.cache.set(link, { indicator: indicator.start(link, indicators), url: newUrl, settings })
            } else {
                this.cache.set(link, { indicator: undefined, url: newUrl, settings })
            }
            indicator.end(prevIndicator)
        });
    }

    private async update(url: URL): Promise<void> {
        try {
            const r = await fetch(this.urlToStr(url), {
                method: "GET",
                headers: { "D00r": this.id },
            });
            if (!r.ok) {
                throw new Error("code " + r.status);
            }
            this.activateLinks(url);
        } catch (e) {
            location.reload()
        }
    }

    private urlCurrent(): URL {
        return new URL(window.location.href)
    }
    private urlToStr(url: URL): string {
        return url.pathname + (url.search ? (url.pathname.endsWith("/") ? url.search : "/" + url.search) : "")
    }

    private urlAreEqual(url1: URL, url2: URL) {
        return url1.pathname === url2.pathname && this.searchEqual(this.searchToMap(url1.searchParams), this.searchToMap(url2.searchParams))
    }
    public push(path: string, activate: boolean = true) {
        const newUrl = new URL(path, window.location.origin);
        if (activate) {
            this.activateLinks(newUrl);
        }
        const currentUrl = new URL(this.urlCurrent(), window.location.origin)
        if (!this.urlAreEqual(currentUrl, newUrl) && !detached) {
            history.pushState(null, '', path);
        }
    }
    public replace(path: string): void {
        const newUrl = new URL(path, window.location.origin);
        this.activateLinks(newUrl);
        const currentUrl = new URL(this.urlCurrent(), window.location.origin)
        if (!this.urlAreEqual(currentUrl, newUrl) && !detached) {
            history.replaceState(null, '', path);
            return
        }
    }

}

export default new Navigator(id)
