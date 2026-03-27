// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

if (document.currentScript!.dataset.lic == null) {
	console.info("DOORS AGPL-3.0-only")
}

export const id: string = document.currentScript!.id
export const prefix: string = document.currentScript!.dataset.prefix!
export const rootId: number = Number(document.currentScript!.dataset.root)
export const ttl: number = Number(document.currentScript!.dataset.ttl)
export const disconnectAfter: number = Number(document.currentScript!.dataset.disconnect)
export const requestTimeout: number = Number(document.currentScript!.dataset.request)
export const solitairePing: number = Number(document.currentScript!.dataset.ping)
