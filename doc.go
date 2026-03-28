// Package doors builds server-rendered web apps with typed routing, reactive
// state, and server-handled browser interactions.
//
// Most apps start with [NewRouter], register one or more page models with
// [UseModel], and render dynamic fragments with [Door], [Source], and [Beam].
// Event attrs such as [AClick], [ASubmit], and [ALink] connect DOM events,
// forms, and navigation to Go handlers while still producing regular HTML.
//
// For a guided introduction, see the documents embedded in [DocsFS].
package doors
