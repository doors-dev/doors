import { cases } from "./kit"
import "./tracker_cases"
import "./deck_cases"
import "./package_cases"
import "./indicator_cases"

export function names(): Array<string> {
	return [...cases.keys()]
}

export async function run(name: string): Promise<string> {
	const fn = cases.get(name)
	if (fn == undefined) {
		return `unknown case: ${name}`
	}
	try {
		await fn()
		return ""
	} catch (e: any) {
		return String(e?.message ?? e)
	}
}
