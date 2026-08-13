import { Header, Package, Sync, SyncFrame } from "./package";
import { id, solitaireRoll, prefix, noStream } from "./params"
import { ProgressiveDelay, AbortTimer, ReliableTimer, Result, result } from "./lib"


export type Results = Map<number, Result<any, string>>

export type Lost = Array<number>

type Gap = ([number, number] | [number])
type Gaps = Array<Gap>

export const STRESS_MODE = false

export type Frame = {
	results: Results,
	lost: Lost,
}

interface Controller {
	onPackage(p: Package): void
	ack(frame: Frame): void
	drop(frame: Frame): void
	kill(): void
	suspend(): void
	resetTTL(): void
}

export interface Connector {
	send(frame: Frame): void
	pause(): void
	resume(): void
	resetDelays(): void
}

export function NewConnector(ctrl: Controller): Connector {
	return new Conn(ctrl)
}

class Conn {
	readonly receiverDelay = new ProgressiveDelay()
	readonly senderDelay = new ProgressiveDelay()
	private seq = 0
	private inFlight = new Map<number, Frame>()
	private sender: ReportSender
	private frame: Frame | null = null
	private ts: number = 0
	private pausePromise: Promise<void>
	private pauseCont: (() => void) | null = null
	aborts: Set<AbortTimer> = new Set()

	constructor(private ctrl_: Controller) {
		new PackageReceiver(this, 1)
		this.sender = new ReportSender(this)
		this.pausePromise = new Promise((res) => res())
	}
	resetDelays() {
		this.receiverDelay.reset()
		this.senderDelay.reset()
	}
	resume() {
		if (!this.pauseCont) {
			return
		}
		const cont = this.pauseCont
		this.pauseCont = null
		cont()
	}
	pause() {
		if (this.pauseCont) {
			return
		}
		this.pausePromise = new Promise((res) => {
			this.pauseCont = res
		})
		for (const abort of this.aborts) {
			abort.abort()
		}
	}
	send(frame: Frame) {
		this.merge(frame)
	}
	private merge(frame: Frame) {
		if (this.frame == null) {
			this.frame = frame
			return
		}
		this.frame.lost.push(...frame.lost)
		for (const [key, value] of frame.results.entries()) {
			this.frame.results.set(key, value)
		}
	}
	pauseGuard() {
		return this.pausePromise
	}
	sync(id: number, syncFrame: SyncFrame, flush: boolean) {
		this.ts = syncFrame.time
		if (syncFrame.acks.length == 0) {
			if (flush) {
				this.receiverFlush(id)
			}
			return
		}
		for (const id of syncFrame.acks) {
			const frame = this.inFlight.get(id)
			if (!frame) {
				continue
			}
			this.inFlight.delete(id)
			this.ctrl_.ack(frame)
		}
		for (const [id, frame] of this.inFlight.entries()) {
			if (id < syncFrame.acks[syncFrame.acks.length - 1]) {
				this.inFlight.delete(id)
				this.ctrl_.drop(frame)
			}
		}
		if (flush) {
			this.receiverFlush(id)
		}
	}
	private flush() {
		const frame = this.frame
		this.frame = null
		this.seq += 1
		if (frame != null) {
			this.inFlight.set(this.seq, frame)
		}
		this.sender.send(this.seq, this.ts, frame)
	}
	kill() {
		this.ctrl_.kill()
	}
	suspend() {
		this.ctrl_.suspend()
	}
	resetTTL() {
		this.ctrl_.resetTTL()
	}
	private receiveBoundary_ = 1
	private receiveBoundaryBuffer_ = new Set<number>
	private receivePackageBuffer_: Array<{ id: number, p: Package }> = []
	private receiverPendingFlush = new Set<number>()
	private receiverFlush(id: number) {
		if (id <= this.receiveBoundary_) {
			this.flush()
			return
		}
		this.receiverPendingFlush.add(id)
	}
	receiverDone(id: number, flush: boolean) {
		this.receiveBoundaryBuffer_.add(id)
		if (this.receiveBoundary_ != id) {
			return
		}
		while (this.receiveBoundaryBuffer_.has(this.receiveBoundary_)) {
			this.receiveBoundaryBuffer_.delete(this.receiveBoundary_)
			this.receiveBoundary_ += 1
		}
		const buffer = this.receivePackageBuffer_
		this.receivePackageBuffer_ = []
		for (const pb of buffer) {
			this.receiverPackage(pb.id, pb.p)
		}
		for (const id of [...this.receiverPendingFlush]) {
			if (id <= this.receiveBoundary_) {
				flush = true
				this.receiverPendingFlush.delete(id)
			}
		}
		if (flush) {
			this.flush()
		}
	}
	private _receiverPackage(id: number, p: Package) {
		this.receiverPendingFlush.add(id)
		if (id <= this.receiveBoundary_) {
			this.ctrl_.onPackage(p)
			return
		}
		this.receivePackageBuffer_.push({ id, p })
		return
	}
	receiverPackage(id: number, p: Package) {
		if (STRESS_MODE && Math.random() > 0.5) {
			setTimeout(() => {
				this._receiverPackage(id, p)
			}, Math.round(Math.random() * 200))
			return
		}
		this._receiverPackage(id, p)
		return
	}
}


class ReportSender {
	private streamer: ReportStreamer | null = null
	constructor(private conn_: Conn) {
		if (supportsRequestStreams) {
			this.streamer = new ReportStreamer(this.conn_)
		}
	}
	send(id: number, ts: number, frame: Frame | null) {
		const payload = [id, ts, {}, []]
		if (frame) {
			frame.lost.sort((a, b) => a - b)
			const gaps: Gaps = []
			let gap: any
			for (const seq of frame.lost) {
				if (!gap) {
					gap = [seq]
					continue
				}
				const prev = gap[1] !== undefined ? gap[1] : gap[0]
				if (seq == prev + 1) {
					gap[1] = seq
					continue
				}
				gaps.push(gap)
				gap = [seq]
			}
			if (gap) {
				gaps.push(gap)
			}
			payload[2] = Object.fromEntries(frame.results)
			payload[3] = gaps
		}
		const message = JSON.stringify(payload)
		if (this.streamer) {
			this.streamer.write(message)
			return
		}
		this.request(message)
	}
	private counter = 0
	private async request(body: string) {
		if (STRESS_MODE && Math.random() > 0.5) {
			return
		}
		this.counter += 1
		const abortTimer = new AbortTimer(solitaireRoll)
		const [response, err] = await result(() => {
			return fetch(`${prefix}/s/${id}?t=${this.counter}`, {
				signal: abortTimer.signal,
				method: "POST",
				headers: {
					'Content-Type': 'application/json;charset=UTF-8',
				},
				body: body,
			})
		})
		abortTimer.cancel()
		if (err) {
			return
		}
		if (response.status === 410) {
			this.conn_.kill()
			return
		}
	}
}

class ReportStreamer {
	private messageWriter_: WritableStreamDefaultWriter<Uint8Array>
	private messageReader_: ReadableStreamDefaultReader<Uint8Array>
	private encoder = new TextEncoder()
	constructor(private conn_: Conn) {
		const channel = new TransformStream<Uint8Array, Uint8Array>();
		this.messageReader_ = channel.readable.getReader()
		this.messageWriter_ = channel.writable.getWriter()
		this.launch()
	}
	private launch() {
		const channel = new TransformStream<WritableStreamDefaultWriter, WritableStreamDefaultWriter>();
		const writer = channel.writable.getWriter()
		const reader = channel.readable.getReader()
		new ReportWriter(this.conn_, writer, 1)
		this.loop(reader)
	}
	async write(report: string) {
		const textBytes = this.encoder.encode(report);
		const buffer = new ArrayBuffer(4 + textBytes.length);
		const view = new DataView(buffer);
		view.setUint32(0, textBytes.length);
		new Uint8Array(buffer, 4).set(textBytes);
		await this.messageWriter_.write(new Uint8Array(buffer))
	}
	private async loop(readerForWriter: ReadableStreamDefaultReader) {
		let writer: WritableStreamDefaultWriter | null = null
		let message: Uint8Array | null = null
		let writerPromise: Promise<ReadableStreamReadResult<WritableStreamDefaultWriter>> | null = null
		let messagePromise: Promise<ReadableStreamReadResult<Uint8Array>> | null = null
		let resolved = false
		while (true) {
			if (!writerPromise) {
				resolved = false
				writerPromise = readerForWriter.read()
				writerPromise.then(() => resolved = true)
			}
			if (!writer || resolved) {
				const { value } = await writerPromise
				if (writer) {
					const prevWriter = writer
					result(() => prevWriter.close())
				}
				writerPromise = null
				writer = value!
				continue
			}
			if (message) {
				const [_, err] = await result(() => writer!.write(message))
				if (err) {
					writer = null
					continue
				}
				message = null
				continue
			}
			if (!messagePromise) {
				messagePromise = this.messageReader_.read()
			}
			const source = await Promise.race([
				messagePromise.then(() => "message"),
				writerPromise.then(() => "writer"),
			])
			if (source === "message") {
				const { value } = await messagePromise
				messagePromise = null
				message = value!
			}
		}
	}
}


class ReportWriter {
	private abortTimer_: AbortTimer
	private rollTimer_: ReliableTimer
	private rolled = false
	private stream = new TransformStream()
	private writer = this.stream.writable.getWriter()
	constructor(private conn_: Conn, private writer_: WritableStreamDefaultWriter<WritableStreamDefaultWriter>, private id: number) {
		this.abortTimer_ = new AbortTimer(solitaireRoll * 4 / 3)
		this.rollTimer_ = new ReliableTimer(solitaireRoll, () => {
			this.roll()
		})
		conn_.aborts.add(this.abortTimer_)
		this.loop()
	}
	private roll() {
		if (this.rolled) {
			return
		}
		this.rollTimer_.cancel()
		this.rolled = true
		this.conn_.pauseGuard().then(() => {
			new ReportWriter(this.conn_, this.writer_, this.id + 1)
		})
	}
	private loop() {
		const promise = fetch(`${prefix}/s/${id}?t=${this.id}`, {
			signal: this.abortTimer_.signal,
			method: "POST",
			body: this.stream.readable,
			//@ts-ignore
			duplex: "half",
			headers: {
				'Content-Type': "application/octet-stream",
			},
		})
		this.writer_.write(this.writer)
		promise.then((response) => {
			result(() => this.writer.abort())
			if (response.status === 410) {
				this.kill()
				return
			}
			if (!response.ok) {
				this.error()
				return
			}
			this.roll()
			this.clear()
		}).catch(() => {
			result(() => this.writer.abort())
			this.error()
		})

	}
	private error() {
		this.clear()
		if (this.rolled) {
			return
		}
		if (STRESS_MODE) {
			this.roll()
			return
		}
		this.conn_.senderDelay.wait().then(() => {
			this.roll()
		})
	}
	private kill() {
		this.clear()
		this.conn_.kill()
	}
	private clear() {
		this.abortTimer_.cancel()
		this.rollTimer_.cancel()
		this.conn_.aborts.delete(this.abortTimer_)
	}
}


class PackageReceiver {
	private abortTimer_: AbortTimer
	private rollTimer_: ReliableTimer
	private rolled = false
	private packageReader: PackageReader
	constructor(private conn_: Conn, private id: number) {
		this.abortTimer_ = new AbortTimer(solitaireRoll * 4 / 3)
		conn_.aborts.add(this.abortTimer_)
		this.rollTimer_ = new ReliableTimer(solitaireRoll, () => {
			this.roll()
		})
		this.packageReader = new PackageReader(this.conn_, this.id)
		this.launch()
	}


	private roll() {
		if (this.rolled) {
			return
		}
		this.rolled = true
		this.rollTimer_.cancel()
		this.conn_.pauseGuard().then(() => {
			new PackageReceiver(this.conn_, this.id + 1)
		})
	}
	private async launch() {
		if (await this.loop()) {
			this.clear()
			return
		}
		this.conn_.receiverDone(this.id, false)
	}

	private async loop(): Promise<boolean> {
		const [response, err] = await result(() => {
			return fetch(`${prefix}/s/${id}?t=${this.id}`, {
				signal: this.abortTimer_.signal,
				method: "GET",
				cache: "no-store",
				headers: {
					Accept: "application/octet-stream",
				},
			})
		})
		if (err) {
			this.error()
			return false
		}
		if (response.status === 410) {
			this.kill()
			return false
		}
		if (!response.ok) {
			this.error()
			return false
		}
		this.conn_.resetTTL()
		const reader = response.body!.getReader()
		while (true) {
			const [read, err] = await result(() => reader.read())
			if (err) {
				this.error()
				return false
			}
			if (read.done) {
				this.error()
				return false
			}
			if (STRESS_MODE && Math.random() > 0.5) {
				// result(() => reader.cancel())
				this.error()
				return false
			}
			const [chunkResult, chunkErr] = await result(() => this.packageReader.readChunk(read.value))
			if (chunkErr) {
				console.error("stream decode error", chunkErr)
				this.error()
				return false
			}
			switch (chunkResult) {
				case readResult.cont:
					break
				case readResult.done:
					this.roll()
					result(() => reader.cancel())
					return true
				case readResult.error:
					this.error()
					return false
			}
		}
	}
	private error() {
		this.clear()
		if (this.rolled) {
			return
		}
		if (STRESS_MODE) {
			this.roll()
			return
		}
		this.conn_.receiverDelay.wait().then(() => {
			this.roll()
		})
	}
	private kill() {
		this.clear()
		this.conn_.kill()
	}
	private clear() {
		this.abortTimer_.cancel()
		this.rollTimer_.cancel()
		this.conn_.aborts.delete(this.abortTimer_)
	}
}

const connectorStatus = {
	signal: "signal",
	sync: "sync",
	header: "header",
	payload: "payload",
} as const
type SyncStatus = typeof connectorStatus[keyof typeof connectorStatus];

const signals = {
	sync: 0x00,
	action: 0x01,
	roll: 0x02,
	suspend: 0x03,
	kill: 0x04,
	terminator: 0xFF,
}


const readResult = {
	cont: "cont",
	done: "done",
	error: "error",
} as const
type ReadResult = typeof readResult[keyof typeof readResult];

class PackageReader {
	private status_: SyncStatus = connectorStatus.signal
	private header_: Header | undefined
	private package_: Package | undefined
	private sync_: Sync | undefined
	private roll = false

	constructor(private conn_: Conn, private id: number) {

	}
	async readChunk(data: Uint8Array): Promise<ReadResult> {
		if (data.length == 0) {
			return readResult.cont
		}
		if (this.status_ == connectorStatus.signal) {
			const signal = data[0]
			switch (signal) {
				case signals.sync:
					this.status_ = connectorStatus.sync
					this.sync_ = new Sync()
					if (data.length == 1) {
						return readResult.cont
					}
					return await this.readChunk(data.subarray(1))
				case signals.action:
					this.status_ = connectorStatus.header
					this.header_ = new Header()
					if (data.length == 1) {
						return readResult.cont
					}
					return await this.readChunk(data.subarray(1))
				case signals.suspend:
					this.conn_.suspend()
					return readResult.done
				case signals.kill:
					this.conn_.kill()
					return readResult.done
				case signals.roll:
					this.roll = true
					if (data.length == 1) {
						return readResult.cont
					}
					return await this.readChunk(data.subarray(1))
				case signals.terminator:
					this.conn_.receiverDone(this.id, false)
					return readResult.done
				default:
					console.error(new Error("unsupported signal " + signal))
					return readResult.error
			}
		}
		if (this.status_ == connectorStatus.sync) {
			for (let i = 0; i < data.length; i++) {
				const byte = data[i]
				if (byte != signals.terminator) {
					continue
				}
				this.sync_!.append(data.subarray(0, i))
				const frame = await this.sync_!.frame()
				this.sync_ = undefined
				this.status_ = connectorStatus.signal
				if (this.roll) {
					this.conn_.sync(this.id, frame, false)
					this.conn_.receiverDone(this.id, true)
					return readResult.done
				} else {
					this.conn_.sync(this.id, frame, true)
				}
				if (i + 1 == data.length) {
					return readResult.cont
				}
				return await this.readChunk(data.subarray(i + 1))
			}
			this.sync_!.append(data)
			return readResult.cont
		}
		if (this.status_ == connectorStatus.header) {
			for (let i = 0; i < data.length; i++) {
				const byte = data[i]
				if (byte != signals.terminator) {
					continue
				}
				this.header_!.append(data.subarray(0, i))
				this.package_ = await this.header_!.package()
				this.header_ = undefined
				this.status_ = connectorStatus.payload
				if (await this.package_!.finalize()) {
					this.conn_.receiverPackage(this.id, this.package_)
					this.package_ = undefined
					this.status_ = connectorStatus.signal
				}
				if (i + 1 == data.length) {
					return readResult.cont
				}
				return await this.readChunk(data.subarray(i + 1))
			}
			this.header_!.append(data)
			return readResult.cont
		}
		const remaining = this.package_!.remaining();
		const chunk = remaining >= data.length ? data : data.subarray(0, remaining)
		this.package_!.append(chunk)
		if (await this.package_!.finalize()) {
			this.conn_.receiverPackage(this.id, this.package_!)
			this.package_ = undefined
			this.status_ = connectorStatus.signal
		}
		if (chunk.length == data.length) {
			return readResult.cont
		}
		return this.readChunk(data.subarray(chunk.length))
	}
}

export const supportsRequestStreams = (() => {
	if (noStream) {
		return false
	}
	const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming | undefined;
	const protocol = nav?.nextHopProtocol;
	if (protocol !== "h2" && protocol !== "h3") {
		return false
	}
	let duplexAccessed = false;
	const hasContentType = new Request("", {
		body: new ReadableStream(),
		method: 'POST',
		cache: "no-store",
		//@ts-ignore
		get duplex() {
			duplexAccessed = true;
			return 'half';
		},
	}).headers.has('Content-Type');
	if (!duplexAccessed || hasContentType) {
		return false
	}
	return true
})();
