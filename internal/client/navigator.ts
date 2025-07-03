import { id } from "./params"
import { arraysEqual, splitClass } from "./lib"
import indicator from "./indicator"


type PathMatchType = "full" | "starts" | "parts";
type QueryMatchType = "all" | "some";


export class Navigator {
    private lastActivated: string | null = null;
    private indications = new WeakMap<HTMLElement, number | undefined>()
    constructor(
        private id: string,
    ) {
        window.addEventListener('popstate', async () => {
            await this.update(this.pagePath());
        });

        document.addEventListener("DOMContentLoaded", () => this.activateLinks(this.pagePath()));
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


    public activateInside(e: HTMLElement | DocumentFragment): void {
        this.activateLinks(this.lastActivated!, e)
    }

    private trim(path: string): string {
        return path.replace(/^\/+|\/+$/g, "")
    }

    private activateLinks(path: string, parent: any = document): void {
        if (parent === document) {
            if (path === this.lastActivated) {
                return
            }
            this.lastActivated = path;
        }
        const links = document.querySelectorAll('[data-d00r-active]');
        const newUrl = new URL(path, window.location.origin);
        links.forEach(linkElement => {
            const link = linkElement as HTMLAnchorElement
            const attr = link.getAttribute("data-d00r-active");
            if (!attr) {
                return
            }
            const href = link.getAttribute("href")
            if (href === null) {
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
            const prevIndicator = this.indications.get(link)
            if (match) {
                this.indications.set(link, indicator.start(link, indicators))
            }
            indicator.end(prevIndicator)
            /*
             *
            const activeClasses = splitClass(activeClass)
            const activeRemoveClasses = splitClass(activeRemoveClass)
            if (match) {
                for (const c of activeClasses) {
                    target.classList.add(c);
                }
                for (const c of activeRemoveClasses) {
                    target.classList.remove(c);
                }
                return;
            }
            for (const c of activeClasses) {
                target.classList.remove(c);
            }
            for (const c of activeRemoveClasses) {
                target.classList.add(c);
            } */
        });
    }

    private async update(path: string): Promise<void> {
        try {
            const r = await fetch(path, {
                method: "GET",
                headers: { "D00r": this.id },
            });
            if (!r.ok) {
                throw new Error("code " + r.status);
            }
        } catch (e) {
            console.error(e);
            window.location.href = path;
        }
    }

    private pagePath(): string {
        const path = window.location.pathname;
        const query = window.location.search;
        return path + (query ? (path.endsWith("/") ? query : "/" + query) : "");
    }

    public async forceUpdate(path: string): Promise<void> {
        await this.update(path);
    }

    public push(path: string): void {
        this.activateLinks(path);
        if (window.location.pathname === path) {
            return;
        }
        history.pushState(null, '', path);
    }

    public replace(path: string): void {
        this.activateLinks(path);
        if (window.location.pathname === path) {
            return;
        }
        history.replaceState(null, '', path);
    }

}

export default new Navigator(id)
