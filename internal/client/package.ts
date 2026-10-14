// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


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

export type EncodedPayload = [PayloadType, any]
export type DecodedPayload = Payload | Promise<Payload>

export function decodePayload(payload: EncodedPayload | undefined): DecodedPayload {
	if (payload === undefined) {
		return {}
	}
	const [payloadType, content] = payload
	if (payloadType === payloadTypes.none) {
		return {}
	}
	if (!isGzip(payloadType)) {
		if (isText(payloadType)) {
			return {
				text: content,
				any: content,
			}
		}
		if (isJson(payloadType)) {
			const json = content
			return {
				json,
				any: json,
			}
		}
	}
	return (async () => {
		const prefix = "data:application/octet-stream;base64,"
		const res = await fetch(prefix + content)
		let stream = res.body!
		if (isGzip(payloadType)) {
			const ds = new DecompressionStream("gzip")
			stream = stream.pipeThrough(ds)
		}
		const resp = new Response(stream)
		if (isBinary(payloadType)) {
			const binary = await resp.arrayBuffer()
			return {
				binary,
				any: binary,
			}
		}
		if (isText(payloadType)) {
			const text = await resp.text()
			return {
				text,
				any: text,
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
	})()
}

export class Package {
	group: number
	batch: number
	door: number
	dependency: number
	dependencyDoor: number
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
		if (!Array.isArray(header[0])) {
			this.start = header[0]
			this.end = header[0]
		} else {
			this.start = header[0][0]
			this.end = header[0][1]
		}
		this.isFiller = header.length == 1
		if (this.isFiller) {
			this.group = 0
			this.batch = 0
			this.door = 0
			this.dependency = 0
			this.dependencyDoor = 0
			this.action = ""
			this.arg = undefined
			return
		}
		this.door = header[1][0]
		this.dependency = header[1][1] ?? 0
		this.dependencyDoor = header[1][2] ?? 0
		if (Array.isArray(header[2])) {
			this.group = header[2][0]
			this.batch = header[2][1]
		} else {
			this.group = header[2]
			this.batch = 0
		}
		this.action = header[3][0]
		this.arg = header[3][1]
	}

	private written = 0

	remaining(): number {
		return this.length - this.written
	}

	private stream(): ReadableStream {
		const blob = new Blob(this.parts as any)
		this.parts = []
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


export class SyncFrame {
	time: number
	acks: Array<number> = []
	constructor(tuple: Array<any>) {
		this.time = tuple[0]
		if (tuple.length > 1) {
			this.acks = tuple[1]
			this.acks.sort((a, b) => a - b)
		}
	}
}

export class Sync {
	private parts: Array<Uint8Array> = []
	append(buf: Uint8Array) {
		if (buf.length == 0) {
			return
		}
		this.parts.push(buf)
	}
	async frame(): Promise<SyncFrame> {
		const tuple = await new Response(new Blob(this.parts as any, { type: "application/json" })).json();
		this.parts = []
		return new SyncFrame(tuple)
	}
}

export class Header {
	private parts: Array<Uint8Array> = []

	async package(): Promise<Package> {
		const header = await new Response(new Blob(this.parts as any, { type: "application/json" })).json();
		this.parts = []
		return new Package(header)
	}

	append(buf: Uint8Array) {
		if (buf.length == 0) {
			return
		}
		this.parts.push(buf)
	}
}
