// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial


const payloadTypes = {
	none: 0x00,
	binary: 0x01,
	json: 0x02,
	text: 0x03,
	binaryGz: 0x11,
	jsonGz: 0x12,
	textGz: 0x13,
} as const;

export type PayloadType = typeof payloadTypes[keyof typeof payloadTypes]

function isGzip(payloadType: PayloadType): boolean {
	return (payloadType & 0xF0) == 0x10
}

function isBinary(payloadType: PayloadType): boolean {
	return (payloadType & 0x0F) == 0x01
}

function isJson(payloadType: PayloadType): boolean {
	return (payloadType & 0x0F) == 0x02
}

function isText(payloadType: PayloadType): boolean {
	return (payloadType & 0x0F) == 0x03
}


export type Payload = {
	text?: string,
	binary?: ArrayBuffer,
	json?: any,
	any?: any,
}


export async function decodePayload(payload: [PayloadType, string] | undefined): Promise<Payload> {
	if (payload === undefined) {
		return {}
	}
	const [payloadType, encoded] = payload
	if (payloadType === payloadTypes.none) {
		return {}
	}
	const prefix = "data:application/octet-stream;base64,";
	const res = await fetch(prefix + encoded);
	let stream = res.body!
	if (isGzip(payloadType)) {
		const ds = new DecompressionStream("gzip");
		stream = stream.pipeThrough(ds)
	}
	const resp = new Response(stream)
	if (isText(payloadType)) {
		const text = await resp.text()
		return {
			text,
			any: text,
		}
	}
	if (isBinary(payloadType)) {
		const binary = await resp.arrayBuffer()
		return {
			binary,
			any: binary,
		}
	}
	if (isJson(payloadType)) {
		const json = await resp.json()
		return {
			json,
			any: json,
		}
	}
	throw new Error("unsupported payload type")
}

export class Package {
	start: number
	end: number
	isFiller: boolean
	action: string
	arg: any
	private parts: Uint8Array[] = []
	private length = 0
	private payloadType: PayloadType = payloadTypes.none
	private payload: Payload | undefined = undefined

	constructor(header: Array<any>) {
		const payloadInfo = header.pop()
		this.payloadType = payloadInfo ? payloadInfo[0] : payloadTypes.none
		this.length = payloadInfo ? payloadInfo[1] : 0;
		this.end = header[0][0]
		this.start = header[0].length == 2 ? header[0][1] : header[0][0]
		this.action = header.length == 2 ? header[1][0] : ""
		this.arg = header.length == 2 ? header[1][1] : undefined
		this.isFiller = header.length == 1
	}

	private written = 0

	remaining(): number {
		return this.length - this.written
	}

	private stream(): ReadableStream {
		const blob = new Blob(this.parts as any)
		if (!isGzip(this.payloadType)) {
			return blob.stream()
		}
		const ds = new DecompressionStream("gzip");
		return blob.stream().pipeThrough(ds);
	}

	async finalize(): Promise<boolean> {
		if (this.payload != undefined) {
			throw new Error("already finalized")
		}
		if (this.remaining() != 0) {
			return false
		}
		if (this.payloadType == payloadTypes.none) {
			this.payload = {}
			return true
		}
		const resp = new Response(this.stream())
		if (isText(this.payloadType)) {
			let text = ""
			if (this.length != 0) {
				text = await resp.text()
			}
			this.payload = {
				text,
				any: text,
			}
			return true
		}
		if (isBinary(this.payloadType)) {
			let binary = new ArrayBuffer(0)
			if (this.length != 0) {
				binary = await resp.arrayBuffer()
			}
			this.payload = {
				binary,
				any: binary,
			}
			return true
		}
		if (isJson(this.payloadType)) {
			let json = null
			if (this.length != 0) {
				json = await resp.json()
			}
			this.payload = {
				json,
				any: json,
			}
			return true
		}
		throw new Error("unsupported payload type")
	}

	append(buf: Uint8Array) {
		if (this.payload != undefined) {
			throw new Error("payload already finalized")
		}
		if (buf.length > this.remaining()) {
			throw new Error("overflow")
		}
		this.parts.push(buf)
		this.written += buf.length
	}

	getPayload(): Payload {
		if (this.payload == undefined) {
			throw new Error("payload not finalized")
		}
		return this.payload
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
