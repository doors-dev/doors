// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

export const id: string = document.currentScript!.id
export const rootId: number = Number(document.currentScript!.dataset.root)
export const ttl: number = Number(document.currentScript!.dataset.ttl)
export const disconnectAfter: number = Number(document.currentScript!.dataset.disconnect)
export const requestTimeout: number = Number(document.currentScript!.dataset.request)
export const solitairePing: number = Number(document.currentScript!.dataset.ping)
export const detached: boolean = document.currentScript!.dataset.detached !== undefined
