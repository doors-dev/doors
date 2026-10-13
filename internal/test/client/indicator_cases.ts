import { register, eq } from "./kit"
import indicator, { IndicatorEntry } from "../../client/indicator"

function mount(html: string): Element {
	document.getElementById("indicator-host")?.remove()
	const host = document.createElement("div")
	host.id = "indicator-host"
	host.innerHTML = html
	document.body.append(host)
	return host
}

function look(el: Element, ...attrs: Array<string>): Array<any> {
	return [[...el.classList].sort().join(" "), el.innerHTML, ...attrs.map(name => el.getAttribute(name))]
}

function on(kind: string, param1: string, param2?: string | null): IndicatorEntry {
	return [["target", null], kind as any, param1, param2]
}

register("indicator: a single indicator applies attr, class, remove_class and content to the target and end restores all of them", () => {
	const el = mount(`<button class="a b" data-x="0">go</button>`).firstElementChild!
	const id = indicator.start(el, [
		on("attr", "data-x", "1"),
		on("attr", "disabled", ""),
		on("class", "c  d"),
		on("remove_class", "a"),
		on("content", "<i>wait</i>"),
	])
	eq(look(el, "data-x", "disabled"), ["b c d", "<i>wait</i>", "1", ""])
	indicator.end(id)
	eq(look(el, "data-x", "disabled"), ["a b", "go", "0", null])
})

register("indicator: an attr entry with a null value removes the attribute and end puts the original value back", () => {
	const el = mount(`<a href="/x">go</a>`).firstElementChild!
	const id = indicator.start(el, [on("attr", "href", null)])
	eq(el.hasAttribute("href"), false)
	indicator.end(id)
	eq(el.getAttribute("href"), "/x")
})

register("indicator: entries for one element merge into one indication where the last attr value and the last content win", () => {
	const el = mount(`<p>text</p>`).firstElementChild!
	const id = indicator.start(el, [
		on("attr", "data-x", "1"),
		on("content", "one"),
		on("class", "a"),
		on("attr", "data-x", "2"),
		on("content", "two"),
		on("class", "b"),
	])
	eq(look(el, "data-x"), ["a b", "two", "2"])
	indicator.end(id)
	eq(look(el, "data-x"), ["", "text", null])
})

register("indicator: start returns undefined and changes nothing for null or empty entries and when no element matches", () => {
	const host = mount(`<p>text</p>`)
	eq(indicator.start(host, null), undefined)
	eq(indicator.start(host, []), undefined)
	eq(indicator.start(null, [on("class", "a")]), undefined)
	eq(indicator.start(host, [[["query", "#indicator-host .missing"], "class", "a", null]]), undefined)
	eq(indicator.start(null, [[["parent_query", "div"], "class", "a", null]]), undefined)
	indicator.end(undefined)
	eq(host.outerHTML, `<div id="indicator-host"><p>text</p></div>`)
})

register("indicator: the query selector indicates only the first match and works without a target", () => {
	const host = mount(`<p class="i">1</p><p class="i">2</p>`)
	const id = indicator.start(null, [[["query", "#indicator-host .i"], "class", "on", null]])
	eq([...host.children].map(el => el.className), ["i on", "i"])
	indicator.end(id)
	eq([...host.children].map(el => el.className), ["i", "i"])
})

register("indicator: the query_all selector indicates every match", () => {
	const host = mount(`<p class="i">1</p><p class="i">2</p><p>3</p>`)
	const id = indicator.start(null, [[["query_all", "#indicator-host .i"], "class", "on", null]])
	eq([...host.children].map(el => el.className), ["i on", "i on", ""])
	indicator.end(id)
	eq([...host.children].map(el => el.className), ["i", "i", ""])
})

register("indicator: the parent_query selector indicates the closest matching ancestor and never the target itself", () => {
	const host = mount(`<section class="box"><div class="box"><span class="box">x</span></div></section>`)
	const outer = host.firstElementChild!
	const inner = outer.firstElementChild!
	const target = inner.firstElementChild!
	const id = indicator.start(target, [[["parent_query", ".box"], "class", "on", null]])
	eq([outer.className, inner.className, target.className], ["box", "box on", "box"])
	indicator.end(id)
	eq([outer.className, inner.className, target.className], ["box", "box", "box"])
})

register("indicator: one indicator can address the target and other elements with different selectors", () => {
	const host = mount(`<button>go</button><p class="i">1</p>`)
	const button = host.firstElementChild!
	const id = indicator.start(button, [
		on("content", "wait"),
		[["query", "#indicator-host .i"], "attr", "hidden", ""],
	])
	eq([button.innerHTML, host.lastElementChild!.hasAttribute("hidden")], ["wait", true])
	indicator.end(id)
	eq([button.innerHTML, host.lastElementChild!.hasAttribute("hidden")], ["go", false])
})

register("indicator: a second indicator on a busy element queues unapplied, takes over when the first ends, and the element is original after both end", () => {
	const el = mount(`<button class="base" data-x="0">go</button>`).firstElementChild!
	const first = indicator.start(el, [on("class", "one"), on("attr", "data-x", "1"), on("content", "first")])
	const second = indicator.start(el, [on("class", "two"), on("attr", "data-y", "2")])
	eq(look(el, "data-x", "data-y"), ["base one", "first", "1", null])
	indicator.end(first)
	eq(look(el, "data-x", "data-y"), ["base two", "go", "0", "2"])
	indicator.end(second)
	eq(look(el, "data-x", "data-y"), ["base", "go", "0", null])
})

register("indicator: ending the queued indicator before the active one drops it without touching the element, and the active end restores the original", () => {
	const el = mount(`<button class="base" data-x="0">go</button>`).firstElementChild!
	const first = indicator.start(el, [on("class", "one"), on("attr", "data-x", "1"), on("content", "first")])
	const second = indicator.start(el, [on("class", "two"), on("attr", "data-y", "2")])
	indicator.end(second)
	eq(look(el, "data-x", "data-y"), ["base one", "first", "1", null])
	indicator.end(first)
	eq(look(el, "data-x", "data-y"), ["base", "go", "0", null])
})

register("indicator: queued indicators take over in start order", () => {
	const el = mount(`<p>text</p>`).firstElementChild!
	const ids = ["one", "two", "three"].map(name => indicator.start(el, [on("class", name)]))
	eq(el.className, "one")
	indicator.end(ids[0])
	eq(el.className, "two")
	indicator.end(ids[1])
	eq(el.className, "three")
	indicator.end(ids[2])
	eq(el.className, "")
})

register("indicator: an indicator ended while queued is skipped by the handover", () => {
	const el = mount(`<p>text</p>`).firstElementChild!
	const ids = ["one", "two", "three"].map(name => indicator.start(el, [on("class", name)]))
	indicator.end(ids[1])
	eq(el.className, "one")
	indicator.end(ids[0])
	eq(el.className, "three")
	indicator.end(ids[2])
	eq(el.className, "")
})

register("indicator: handover keeps what both indications share, swaps what differs, and the last end restores the value saved before the first", () => {
	const el = mount(`<button class="base gone" data-x="0" data-z="z">go</button>`).firstElementChild!
	const first = indicator.start(el, [
		on("class", "busy one"),
		on("remove_class", "gone"),
		on("attr", "data-x", "1"),
		on("attr", "data-y", "y"),
		on("attr", "data-z", null),
	])
	const second = indicator.start(el, [
		on("class", "busy two"),
		on("remove_class", "gone base"),
		on("attr", "data-x", "2"),
		on("attr", "data-y", "y"),
	])
	eq(look(el, "data-x", "data-y", "data-z"), ["base busy one", "go", "1", "y", null])
	indicator.end(first)
	eq(look(el, "data-x", "data-y", "data-z"), ["busy two", "go", "2", "y", "z"])
	indicator.end(second)
	eq(look(el, "data-x", "data-y", "data-z"), ["base gone", "go", "0", null, "z"])
})

register("indicator: handover from a removed class set to an indication without it puts the class back", () => {
	const el = mount(`<p class="a b">text</p>`).firstElementChild!
	const first = indicator.start(el, [on("remove_class", "a b")])
	const second = indicator.start(el, [on("remove_class", "b")])
	eq(el.className, "")
	indicator.end(first)
	eq(el.className, "a")
	indicator.end(second)
	eq([...el.classList].sort().join(" "), "a b")
})

register("indicator: content handover replaces differing content, restores the original for an indication without content, and saves it again for the next", () => {
	const el = mount(`<p><b>text</b></p>`).firstElementChild!
	const ids = [
		indicator.start(el, [on("content", "one")]),
		indicator.start(el, [on("content", "two")]),
		indicator.start(el, [on("class", "plain")]),
		indicator.start(el, [on("content", "four")]),
	]
	eq(el.innerHTML, "one")
	indicator.end(ids[0])
	eq(el.innerHTML, "two")
	indicator.end(ids[1])
	eq(look(el), ["plain", "<b>text</b>"])
	indicator.end(ids[2])
	eq(look(el), ["", "four"])
	indicator.end(ids[3])
	eq(look(el), ["", "<b>text</b>"])
})

register("indicator: empty original content is restored after a content indication", () => {
	const el = mount(`<p></p>`).firstElementChild!
	const id = indicator.start(el, [on("content", "wait")])
	eq(el.innerHTML, "wait")
	indicator.end(id)
	eq(el.innerHTML, "")
})

register("indicator: an indicator spanning two elements is queued on the busy one and active on the free one, each element handing over on its own", () => {
	const host = mount(`<p class="i">1</p><p class="i">2</p>`)
	const [busy, free] = [...host.children]
	const first = indicator.start(busy, [on("class", "one")])
	const both = indicator.start(null, [[["query_all", "#indicator-host .i"], "class", "both", null]])
	eq([busy.className, free.className], ["i one", "i both"])
	indicator.end(first)
	eq([busy.className, free.className], ["i both", "i both"])
	indicator.end(both)
	eq([busy.className, free.className], ["i", "i"])
})

register("indicator: ending an id twice or an unknown id is ignored and does not end a later indicator on the element", () => {
	const el = mount(`<p>text</p>`).firstElementChild!
	const first = indicator.start(el, [on("class", "one")])
	indicator.end(first)
	const second = indicator.start(el, [on("class", "two")])
	indicator.end(first)
	indicator.end(-1)
	eq(el.className, "two")
	indicator.end(second)
	eq(el.className, "")
})

register("indicator: an element freed by its last end starts the next indicator from its current state", () => {
	const el = mount(`<p data-x="0">text</p>`).firstElementChild!
	indicator.end(indicator.start(el, [on("attr", "data-x", "1")]))
	el.setAttribute("data-x", "5")
	const id = indicator.start(el, [on("attr", "data-x", "2")])
	eq(el.getAttribute("data-x"), "2")
	indicator.end(id)
	eq(el.getAttribute("data-x"), "5")
})

register("indicator: an element detached when its indicator ends is left as indicated and forgotten, so a later indicator treats that state as original", () => {
	const host = mount(`<p class="base">text</p>`)
	const el = host.firstElementChild!
	const first = indicator.start(el, [on("class", "one"), on("content", "wait")])
	const queued = indicator.start(el, [on("class", "two")])
	el.remove()
	indicator.end(first)
	eq(look(el), ["base one", "wait"])
	host.append(el)
	const third = indicator.start(el, [on("class", "three")])
	eq(look(el), ["base one three", "wait"])
	indicator.end(queued)
	eq(look(el), ["base one three", "wait"])
	indicator.end(third)
	eq(look(el), ["base one", "wait"])
})
