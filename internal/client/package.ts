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
	start: number
	end: number
	isFiller: boolean
	action: string
	arg: any
	private parts_: Uint8Array[] = []
	private length_ = 0
	private payloadType_: PayloadType = payloadTypes.none
	private payload_: Payload | undefined = undefined

	constructor(header: Array<any>) {
		const payloadInfo = header.pop()
		this.payloadType_ = payloadInfo ? payloadInfo[0] : payloadTypes.none
		this.length_ = payloadInfo ? payloadInfo[1] : 0;
		if (!Array.isArray(header[0])) {
			this.start = header[0]
			this.end = header[0]
		} else {
			this.start = header[0][0]
			this.end = header[0][1]
		}
		this.action = header.length == 2 ? header[1][0] : ""
		this.arg = header.length == 2 ? header[1][1] : undefined
		this.isFiller = header.length == 1
	}

	private written_ = 0

	remaining(): number {
		return this.length_ - this.written_
	}

	private stream(): ReadableStream {
		const blob = new Blob(this.parts_ as any)
		this.parts_ = []
		if (!isGzip(this.payloadType_)) {
			return blob.stream()
		}
		const ds = new DecompressionStream("gzip");
		return blob.stream().pipeThrough(ds);
	}

	async finalize(): Promise<boolean> {
		if (this.payload_ != undefined) {
			throw new Error("already finalized")
		}
		if (this.remaining() != 0) {
			return false
		}
		if (this.payloadType_ == payloadTypes.none) {
			this.payload_ = {}
			return true
		}
		const resp = new Response(this.stream())
		if (isText(this.payloadType_)) {
			let text = ""
			if (this.length_ != 0) {
				text = await resp.text()
			}
			this.payload_ = {
				text,
				any: text,
			}
			return true
		}
		if (isBinary(this.payloadType_)) {
			let binary = new ArrayBuffer(0)
			if (this.length_ != 0) {
				binary = await resp.arrayBuffer()
			}
			this.payload_ = {
				binary,
				any: binary,
			}
			return true
		}
		if (isJson(this.payloadType_)) {
			let json = null
			if (this.length_ != 0) {
				json = await resp.json()
			}
			this.payload_ = {
				json,
				any: json,
			}
			return true
		}
		throw new Error("unsupported payload type")
	}

	append(buf: Uint8Array) {
		if (this.payload_ != undefined) {
			throw new Error("payload already finalized")
		}
		if (buf.length > this.remaining()) {
			throw new Error("overflow")
		}
		this.parts_.push(buf)
		this.written_ += buf.length
	}

	getPayload(): Payload {
		if (this.payload_ == undefined) {
			throw new Error("payload not finalized")
		}
		return this.payload_
	}

}

export class Header {
	private headerParts_: Array<Uint8Array> = []

	async package(): Promise<Package> {
		const header = await new Response(new Blob(this.headerParts_ as any, { type: "application/json" })).json();
		this.headerParts_ = []
		return new Package(header)
	}

	append(buf: Uint8Array) {
		if (buf.length == 0) {
			return
		}
		this.headerParts_.push(buf)
	}
}
