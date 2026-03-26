# Core Concepts

Recommended to read and understand. But skippable.

## Session

Represents a user browser session. Identified via a session ID stored in a cookie.

- By default, the cookie is session-only (removed on browser exit).
- On the server, the session persists as long as at least one page **instance** is active.

## Instance

A live page within a session. Each instance exists in server memory and encapsulates the rendering process, dynamic **door** tree, **hook** mappings, and client sync control.

- Default limit: one session can hold up to 6 active instances (the least active are suspended).
- Instances remain alive for a TTL value after losing connection to the client.

## Door

A dynamic placeholder or container that can be updated, removed, or replaced. It has no visual footprint but enables reactive updates.

- Doors form a synchronized dynamic tree during rendering.
- Features like append/prepend are implemented using placeholder doors that can be replaced.

## Path Model

A typed structure representing the current page route, including query parameters. Defined using struct tags for path variants and bindings.

- Page routing occurs by deserializing the URL into a path model.
- Changing parameters within the same model updates the current instance reactively.
- Switching between different path structs triggers a full instance change.
- Use `SourceBeam` (provided to the page render function) to observe path changes and trigger rerenders.

## Beam / SourceBeam

Composable reactive state primitives.

- `SourceBeam`: holds the original mutable value.
- `Beam`: a derived or observed value; supports subscriptions and reactive rendering.
- Changes propagate top-down, guaranteeing a consistent state during render.
- `SourceBeam` implements the `Beam` interface.

## Attributes

Attributes are used to attach framework-connected behaviors to HTML elements.

- Includes event bindings (hooks), data attributes, etc.
- Enables integration of client/server behavior.

## Hook

An HTTP handler dynamically routed by the framework, typically bound to a DOM event. Initialized as an attribute. 

- Hooks have their own lifecycle, scoped to the dynamic door in which they were created.
- `src`/`href` attributes can route to hook endpoints for secure file serving.
- You can define custom hooks to trigger from JavaScript.

## Context

The standard Go `context.Context`, extended and used throughout *doors* to:

- Identify the **location in the dynamic tree**
- Bind **hooks**
- Read **beam** values
- Track **hook**-triggered changes completions
- Access and modify **instance/session** entities

