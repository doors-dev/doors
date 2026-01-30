// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial


export class Package {
	start: number
	end: number
	payload: ArrayBuffer
	isFiller: boolean
	action: string
	arg: any

	constructor(header: Array<any>) {
		const payloadLength = header.pop()
		this.end = header[0][0]
		this.start = header[0].length == 2 ? header[0][1] : header[0][0]
		this.action = header.length == 2 ? header[1][0] : ""
		this.arg = header.length == 2 ? header[1][1] : undefined
		this.isFiller = header.length == 1
		this.payload = new ArrayBuffer(payloadLength)
		this.view = new Uint8Array(this.payload)
	}

	private written = 0
	private view: Uint8Array

	remaining(): number {
		return this.payload.byteLength - this.written
	}

	append(buf: Uint8Array) {
		if(buf.length > this.remaining()) {
			throw new Error("overflow")
		}
		this.view.set(buf, this.written)
		this.written += buf.length
	}

}

export class Header {
	private headerParts: Array<Uint8Array> = []

	async package(): Promise<Package> {
		const header = await new Response(new Blob(this.headerParts as any, { type: "application/json" })).json();
		return new Package(header)
	}

	append(buf: Uint8Array) {
		if (buf.length == 0) {
			return
		}
		this.headerParts.push(buf)
	}
}
