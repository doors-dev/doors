import { register, eq, fails } from "./kit"
import { Package, Header, Sync, decodePayload } from "../../client/package"

const none = 0x00
const binary = 0x01
const json = 0x02
const text = 0x03
const binaryGz = 0x11
const jsonGz = 0x12
const textGz = 0x13

const utf8 = new TextEncoder()

async function card(line: string, cuts: Array<number> = []): Promise<Package> {
	const bytes = utf8.encode(line + "\n")
	const header = new Header()
	let from = 0
	for (const cut of [...cuts, bytes.length]) {
		header.append(bytes.subarray(from, cut))
		from = cut
	}
	return header.package()
}

function fields(p: Package): Array<any> {
	return [p.start, p.end, p.group, p.batch, p.door, p.dependency, p.dependencyDoor, p.action, p.arg, p.isFiller]
}

async function gzip(data: Uint8Array): Promise<Uint8Array> {
	const stream = new Blob([data as any]).stream().pipeThrough(new CompressionStream("gzip"))
	return new Uint8Array(await new Response(stream).arrayBuffer())
}

function base64(data: Uint8Array): string {
	return btoa(String.fromCharCode(...data))
}

function bytes(buf: ArrayBuffer | undefined): Array<number> {
	return [...new Uint8Array(buf!)]
}

async function deliver(kind: number, body: Uint8Array, size: number): Promise<Package> {
	const p = await card(`[1,[7],0,["test",[1]],[${kind},${body.length}]]`)
	eq(p.remaining(), body.length)
	for (let from = 0; from < body.length; from += size) {
		eq(await p.finalize(), false)
		const chunk = body.subarray(from, from + size)
		p.append(chunk)
		eq(p.remaining(), body.length - from - chunk.length)
	}
	eq(await p.finalize(), true)
	return p
}

const sample = "héllo wörld ✓ <b>&</b>"
const blob = Uint8Array.from({ length: 300 }, (_, i) => (i * 7) % 256)

register("package: single seq with a [T] link and a 0 batch slot sets start and end to the seq and defaults everything else to 0", async () => {
	const p = await card(`[4,[7],0,["test",[1]],null]`)
	eq(fields(p), [4, 4, 0, 0, 7, 0, 0, "test", [1], false])
})

register("package: a [T,S] link carries the dependency, and a single number in the batch slot is the group", async () => {
	const p = await card(`[4,[7,3],5,["test",[1]],null]`)
	eq(fields(p), [4, 4, 5, 0, 7, 3, 0, "test", [1], false])
})

register("package: a [T,S,D] link with zero S carries only the dependency door, and a [G,B] batch slot carries group and batch", async () => {
	const p = await card(`[4,[7,0,9],[5,2],["test",[1]],null]`)
	eq(fields(p), [4, 4, 5, 2, 7, 0, 9, "test", [1], false])
})

register("package: a [beg,end] seq range with a full link and a [0,B] batch slot decodes a batch without a group", async () => {
	const p = await card(`[[2,4],[7,3,9],[0,2],["test",[1]],null]`)
	eq(fields(p), [2, 4, 0, 2, 7, 3, 9, "test", [1], false])
})

register("package: a bare single seq filler has a zeroed link, empty action and no arg", async () => {
	const p = await card(`[4,null]`)
	eq(fields(p), [4, 4, 0, 0, 0, 0, 0, "", undefined, true])
	eq(p.arg === undefined, true)
})

register("package: a bare range filler keeps its range and has a zeroed link", async () => {
	const p = await card(`[[2,4],null]`)
	eq(fields(p), [2, 4, 0, 0, 0, 0, 0, "", undefined, true])
})

register("package: a filler finalizes at once with an empty payload", async () => {
	const p = await card(`[[2,4],null]`)
	eq(p.remaining(), 0)
	eq(await p.finalize(), true)
	eq(p.getPayload(), {})
})

register("package: action name and a nested arg list are taken from the invocation slot unescaped", async () => {
	const p = await card(`[1,[0],0,["location_assign",["/a?b=<c>&d",true,{"k":[1,null]},"ü"]],null]`)
	eq(p.action, "location_assign")
	eq(p.arg, ["/a?b=<c>&d", true, { k: [1, null] }, "ü"])
})

register("package: header bytes split across chunks, inside a multibyte character and around an empty chunk, decode the same", async () => {
	const line = `[[2,4],[7,3,9],[6,5],["test",["✓"]],[3,12]]`
	const at = utf8.encode(line.slice(0, line.indexOf("✓"))).length
	const p = await card(line, [1, 1, at + 1, at + 2])
	eq(fields(p), [2, 4, 6, 5, 7, 3, 9, "test", ["✓"], false])
	eq(p.remaining(), 12)
})

register("package: a header without payload info has nothing remaining and finalizes with an empty payload", async () => {
	const p = await card(`[4,[7],0,["test",[1]],null]`)
	eq(p.remaining(), 0)
	eq(await p.finalize(), true)
	eq(p.getPayload(), {})
})

register("package: an explicit none payload type finalizes with an empty payload", async () => {
	const p = await card(`[4,[7],0,["test",[1]],[${none},0]]`)
	eq(await p.finalize(), true)
	eq(p.getPayload(), {})
})

register("package: a text payload delivered whole decodes to text and any", async () => {
	const p = await deliver(text, utf8.encode(sample), 1000)
	eq(p.getPayload(), { text: sample, any: sample })
})

register("package: a text payload delivered byte by byte counts down remaining and decodes multibyte characters", async () => {
	const p = await deliver(text, utf8.encode(sample), 1)
	eq(p.getPayload(), { text: sample, any: sample })
})

register("package: a json payload delivered whole decodes to json and any", async () => {
	const p = await deliver(json, utf8.encode(`{"a":[1,2,{"b":"ü"}],"c":null}\n`), 1000)
	eq(p.getPayload(), { json: { a: [1, 2, { b: "ü" }], c: null }, any: { a: [1, 2, { b: "ü" }], c: null } })
})

register("package: a json payload delivered in several chunks finalizes only after the last one", async () => {
	const p = await deliver(json, utf8.encode(`{"a":[1,2,{"b":"ü"}],"c":null}\n`), 7)
	eq(p.getPayload().json, { a: [1, 2, { b: "ü" }], c: null })
})

register("package: a binary payload delivered whole keeps every byte", async () => {
	const p = await deliver(binary, blob, 1000)
	eq(bytes(p.getPayload().binary), [...blob])
	eq(p.getPayload().any === p.getPayload().binary, true)
})

register("package: a binary payload delivered in several chunks keeps byte order", async () => {
	const p = await deliver(binary, blob, 64)
	eq(bytes(p.getPayload().binary), [...blob])
})

register("package: a gzipped text payload counts compressed bytes and decodes to the original text", async () => {
	const packed = await gzip(utf8.encode(sample))
	eq((await deliver(textGz, packed, 1000)).getPayload(), { text: sample, any: sample })
	eq((await deliver(textGz, packed, 5)).getPayload(), { text: sample, any: sample })
})

register("package: a gzipped json payload decodes whole and in chunks", async () => {
	const packed = await gzip(utf8.encode(`{"a":[1,2,3]}\n`))
	eq((await deliver(jsonGz, packed, 1000)).getPayload().json, { a: [1, 2, 3] })
	eq((await deliver(jsonGz, packed, 5)).getPayload().json, { a: [1, 2, 3] })
})

register("package: a gzipped binary payload decodes whole and in chunks", async () => {
	const packed = await gzip(blob)
	eq(bytes((await deliver(binaryGz, packed, 1000)).getPayload().binary), [...blob])
	eq(bytes((await deliver(binaryGz, packed, 5)).getPayload().binary), [...blob])
})

register("package: a zero length text payload finalizes to an empty string", async () => {
	const p = await deliver(text, new Uint8Array(0), 1)
	eq(p.getPayload(), { text: "", any: "" })
})

register("package: a zero length binary payload finalizes to an empty buffer", async () => {
	const p = await deliver(binary, new Uint8Array(0), 1)
	eq(p.getPayload().binary instanceof ArrayBuffer, true)
	eq(p.getPayload().binary!.byteLength, 0)
})

register("package: a zero length json payload finalizes to null", async () => {
	const p = await deliver(json, new Uint8Array(0), 1)
	eq(p.getPayload().json, null)
	eq("json" in p.getPayload(), true)
})

register("package: a chunk larger than remaining throws overflow and is not counted", async () => {
	const body = utf8.encode("hello")
	const p = await card(`[1,[7],0,["test",[1]],[${text},${body.length}]]`)
	p.append(body.subarray(0, 2))
	await fails(() => p.append(body.subarray(1)), "overflow")
	eq(p.remaining(), 3)
	p.append(body.subarray(2))
	eq(await p.finalize(), true)
	eq(p.getPayload().text, "hello")
})

register("package: any chunk for a package without payload overflows", async () => {
	const p = await card(`[4,[7],0,["test",[1]],null]`)
	await fails(() => p.append(utf8.encode("x")), "overflow")
	eq(await p.finalize(), true)
})

register("package: a payload shorter than its declared length stays unfinalized and unreadable", async () => {
	const p = await card(`[1,[7],0,["test",[1]],[${text},5]]`)
	p.append(utf8.encode("hell"))
	eq(await p.finalize(), false)
	eq(await p.finalize(), false)
	eq(p.remaining(), 1)
	await fails(() => p.getPayload(), "payload not finalized")
})

register("package: a json payload cut short by its declared length rejects finalize", async () => {
	const p = await card(`[1,[7],0,["test",[1]],[${json},5]]`)
	p.append(utf8.encode(`{"a":[1,2,3]}`).subarray(0, 5))
	await fails(() => p.finalize())
	await fails(() => p.getPayload(), "payload not finalized")
})

register("package: a gzipped payload that is not gzip rejects finalize", async () => {
	const p = await card(`[1,[7],0,["test",[1]],[${textGz},5]]`)
	p.append(utf8.encode("hello"))
	await fails(() => p.finalize())
})

register("package: an unknown payload type rejects finalize as unsupported", async () => {
	const p = await card(`[1,[7],0,["test",[1]],[4,1]]`)
	p.append(utf8.encode("x"))
	await fails(() => p.finalize(), "unsupported payload type")
})

register("package: finalize after success throws already finalized and append throws payload already finalized", async () => {
	const p = await deliver(text, utf8.encode("hi"), 2)
	await fails(() => p.finalize(), "already finalized")
	await fails(() => p.append(new Uint8Array(0)), "payload already finalized")
	eq(p.getPayload().text, "hi")
})

register("package: decodePayload returns an empty payload synchronously for undefined and for the none type", () => {
	eq(decodePayload(undefined), {})
	eq(decodePayload([none, null]), {})
})

register("package: decodePayload returns inline text and json synchronously", () => {
	eq(decodePayload([text, sample]), { text: sample, any: sample })
	eq(decodePayload([json, { a: [1, 2] }]), { json: { a: [1, 2] }, any: { a: [1, 2] } })
	eq(decodePayload([json, null]), { json: null, any: null })
})

register("package: decodePayload decodes base64 binary asynchronously", async () => {
	const decoded = decodePayload([binary, base64(blob)])
	eq(decoded instanceof Promise, true)
	const payload = await decoded
	eq(bytes(payload.binary), [...blob])
	eq(payload.any === payload.binary, true)
})

register("package: decodePayload decodes an empty base64 binary to an empty buffer", async () => {
	eq(bytes((await decodePayload([binary, ""])).binary), [])
})

register("package: decodePayload inflates base64 gzipped text, json and binary", async () => {
	eq(await decodePayload([textGz, base64(await gzip(utf8.encode(sample)))]), { text: sample, any: sample })
	eq(await decodePayload([jsonGz, base64(await gzip(utf8.encode(`{"a":[1,2,3]}\n`)))]), { json: { a: [1, 2, 3] }, any: { a: [1, 2, 3] } })
	eq(bytes((await decodePayload([binaryGz, base64(await gzip(blob))])).binary), [...blob])
})

register("package: decodePayload rejects an unknown payload type as unsupported", async () => {
	await fails(() => decodePayload([4 as any, ""]), "unsupported payload type")
})

register("package: a sync frame with only a time has no acks", async () => {
	const sync = new Sync()
	sync.append(utf8.encode(`[1700000000123]`))
	const frame = await sync.frame()
	eq([frame.time, frame.acks], [1700000000123, []])
})

register("package: a sync frame split across chunks sorts its acks ascending and the reader is reusable", async () => {
	const sync = new Sync()
	sync.append(utf8.encode(`[1700000000123,[10,`))
	sync.append(new Uint8Array(0))
	sync.append(utf8.encode(`2,33,4]]`))
	const frame = await sync.frame()
	eq([frame.time, frame.acks], [1700000000123, [2, 4, 10, 33]])
	sync.append(utf8.encode(`[5]`))
	eq((await sync.frame()).time, 5)
})
