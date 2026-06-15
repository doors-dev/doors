export const hookErrKinds = {
	canceled: "canceled",
	gone: "gone",
	not_found: "not_found",
	other: "other",
	network: "network",
	bad_request: "bad_request",
	server: "server",
} as const

type HookErrKind = typeof hookErrKinds[keyof typeof hookErrKinds]

export class HookErr extends Error {
	public arg: any
	public status: number | undefined = undefined
	public cause: Error | undefined = undefined
	constructor(public kind: HookErrKind, opt?: any) {
		let message: string
		switch (kind) {
			case hookErrKinds.not_found:
				message = `hook not found on server, may be done`
				break
			case hookErrKinds.canceled:
				message = `hook is blocked by scope`
				break
			case hookErrKinds.gone:
				message = `instance is stopped`
				break
			case hookErrKinds.other:
				if (opt instanceof Error) {
					message = `other error: ${opt.message}`
				} else if (opt && opt.status) {
					message = `other error: status ${opt.status}`
				} else {
					message = "other error"
				}
				break
			case hookErrKinds.network:
				message = opt?.message
				break
			case hookErrKinds.server:
				message = `server error: ${opt?.status}`
				break
			case hookErrKinds.bad_request:
				message = `body parsing error, bad request`
				break
			default:
				throw new Error(`unsupported error type: ${kind}`)
		}

		const cause = opt instanceof Error ? opt : undefined
		// @ts-expect-error: Error constructor overload not recognized by TS (ES2022 feature)
		super(message, cause ? { cause } : undefined)
		if (opt && opt.status && typeof opt.status == "number") {
			this.status = opt.status
		}
		if (cause) {
			this.cause = cause
		}
	}
	canceled() { return this.kind === hookErrKinds.canceled; }
	notFound() { return this.kind === hookErrKinds.not_found; }
	gone() { return this.kind === hookErrKinds.gone; }
	other() { return this.kind === hookErrKinds.other; }
	network() { return this.kind === hookErrKinds.network; }
	server() { return this.kind === hookErrKinds.server; }
	badRequest() { return this.kind === hookErrKinds.bad_request; }
}

