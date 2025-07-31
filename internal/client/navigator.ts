import { id } from "./params"
import { arraysEqual, splitClass } from "./lib"
import indicator from "./indicator"


type PathMatchType = "full" | "starts" | "parts";
type QueryMatchType = "all" | "some";


type State = {
    indicator: number | undefined,
    url: URL,
}

export class Navigator {
    private state = new WeakMap<Element, State>()
    constructor(
        private id: string,
    ) {
        window.addEventListener('popstate', async () => {
            await this.update(this.urlCurrent());
        });

        document.addEventListener("DOMContentLoaded", () => this.activateLinks(this.urlCurrent()));
    }

    private searchEqual(
        search1: URLSearchParams,
        search2: URLSearchParams,
        params?: string[]
    ): boolean {
        const toObj = (searchParams: URLSearchParams): Record<string, string[]> => {
            const obj: Record<string, string[]> = {};
            for (const key of searchParams.keys()) {
                if (!params || params.length === 0 || params.includes(key)) {
                    obj[key] = searchParams.getAll(key);
                }
            }
            return obj;
        };

        const obj1 = toObj(search1);
        const obj2 = toObj(search2);

        if (!arraysEqual(Object.keys(obj1), Object.keys(obj2), false)) {
            return false;
        }
        for (const key in obj1) {
            if (!arraysEqual(obj1[key], obj2[key])) {
                return false;
            }
        }
        return true;
    }


    public activateInside(e: Element | DocumentFragment): void {
        this.activateLinks(this.urlCurrent(), e)
    }

    private trim(path: string): string {
        return path.replace(/^\/+|\/+$/g, "")
    }

    private activateLinks(newUrl: URL, parent: any = document): void {
        const links = parent.querySelectorAll('[data-d00r-active]');
        links.forEach(linkElement => {
            const state = this.state.get(linkElement)
            const link = linkElement as HTMLAnchorElement
            const attr = link.getAttribute("data-d00r-active");
            if (!attr) {
                return
            }
            const href = link.getAttribute("href")
            if (href === null) {
                return
            }
            if (!!state && this.urlAreEqual(state.url, newUrl)) {
                return
            }
            const [pathMatchTuple, queryMatchTuple, indicators]: any = JSON.parse(attr);
            const pathMatch: PathMatchType = pathMatchTuple[0];
            const pathMatchArg: number | undefined = pathMatchTuple[1];
            const queryMatch: QueryMatchType = queryMatchTuple[0];
            const queryMatchArg: string[] | undefined = queryMatchTuple[1];
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
            if (match && queryMatch === "all") {
                match = this.searchEqual(newUrl.searchParams, url.searchParams);
            } else if (match && queryMatch === "some" && queryMatchArg) {
                match = this.searchEqual(newUrl.searchParams, url.searchParams, queryMatchArg);
            }
            const prevIndicator = state?.indicator
            if (match) {
                this.state.set(link, { indicator: indicator.start(link, indicators), url: newUrl })
            } else {
                this.state.set(link, { indicator: undefined, url: newUrl })
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
        } catch (e) {
            console.error(e);
            window.location.href = this.urlToStr(url);
        }
    }

    private urlCurrent(): URL {
        return new URL(window.location.href)
    }
    private urlToStr(url: URL): string {
        return url.pathname + (url.search ? (url.pathname.endsWith("/") ? url.search : "/" + url.search) : "")
    }

    private urlAreEqual(url1: URL, url2: URL) {
        return url1.pathname === url2.pathname && this.searchEqual(url1.searchParams, url2.searchParams)
    }

    public push(path: string): void {
        const currentUrl = new URL(this.urlCurrent(), window.location.origin)
        const newUrl = new URL(path, window.location.origin);
        if (this.urlAreEqual(currentUrl, newUrl)) {
            return
        }
        this.activateLinks(newUrl);
        history.pushState(null, '', path);
    }
    public replace(path: string): void {
        const currentUrl = new URL(this.urlCurrent(), window.location.origin)
        const newUrl = new URL(path, window.location.origin);
        if (this.urlAreEqual(currentUrl, newUrl)) {
            return
        }
        this.activateLinks(newUrl);
        history.replaceState(null, '', path);
    }

}

export default new Navigator(id)
