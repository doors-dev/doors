// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

const license = document.currentScript!.dataset.license?.split(":")

const agplMsg = "AGPL-3.0-only mode (see LICENSE).";


if (license == null) {
	console.info("[doors] " + agplMsg);
} else {
	const [id, tier, domain] = license
	const isLocalHost = ["localhost", "127.0.0.1", "[::1]"].includes(location.hostname)
	const correctDomain = domain === "*" || location.hostname === domain || location.hostname.endsWith("." + domain);
	if (!isLocalHost && !correctDomain) {
		console.warn(
			[
				"[doors] Invalid license provided.",
				"Id: " + id,
				"Licensed domain: " + domain,
				"AGPL-3.0-only mode (see LICENSE).",
			].join("\n")
		);
	} else {
		console.info(
			[
				"[doors] " + tier + " license provided.",
				...isLocalHost ? ["Running in localhost mode.", "Licensed domain: " + domain] : [],
				"Id: " + id,
			].join("\n")
		);
	}
}

export const id: string = document.currentScript!.id
export const rootId: number = Number(document.currentScript!.dataset.root)
export const ttl: number = Number(document.currentScript!.dataset.ttl)
export const disconnectAfter: number = Number(document.currentScript!.dataset.disconnect)
export const requestTimeout: number = Number(document.currentScript!.dataset.request)
export const solitairePing: number = Number(document.currentScript!.dataset.ping)
export const detached: boolean = document.currentScript!.dataset.detached !== undefined
export const noGzip: boolean = document.currentScript!.dataset.nogzip !== undefined
