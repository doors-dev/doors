// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

const license = document.currentScript!.dataset.license?.split(":")
const purchaseMessage = "For commercial production use, purchase a license at https://doors.dev or via sales@doors.dev."
const ncMessage = "Running in development/non-commercial mode."

if (license == null) {
    console.warn(
        [
            "[doors] No license provided.",
            ncMessage,
            purchaseMessage,
        ].join("\n")
    );
} else {
    const [id, tier, domain] = license
    const isLocalHost = ["localhost", "127.0.0.1", "[::1]"].includes(location.hostname)
    const correctDomain = domain === "*" || location.hostname === domain || location.hostname.endsWith("." + domain);
    if (!isLocalHost && !correctDomain) {
        console.error(
            [
                "[doors] Invalid license provided.",
                ncMessage,
                "Id: " + id,
                "Licensed domain: " + domain,
                purchaseMessage
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
