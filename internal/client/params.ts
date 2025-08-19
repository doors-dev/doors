export const id: string = document.currentScript!.id
export const rootId: number = Number(document.currentScript!.dataset.root)
export const ttl: number = Number(document.currentScript!.dataset.ttl)
export const sleepAfter: number = Number(document.currentScript!.dataset.sleep)
export const requestTimeout: number = Number(document.currentScript!.dataset.request)
export const detached: boolean = document.currentScript!.dataset.detached !== undefined
