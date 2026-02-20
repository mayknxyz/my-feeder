# 003 — Bubbletea + Elm architecture

## Status

Accepted

## Context

A TUI app needs an event loop, input handling, screen rendering, and background task coordination. Common approaches:

1. **Imperative loop** — manual `for { select {} }` with raw terminal calls. Maximum control, maximum boilerplate.
2. **Component-based** (like tview) — widgets with callbacks. Familiar from web UIs but leads to scattered state and callback spaghetti.
3. **Elm architecture** (like bubbletea) — single model, message-driven updates, view as pure function of state. Predictable, testable, no shared mutable state.

## Decision

Use bubbletea with the Elm architecture pattern:

```
Model (all app state)
  → Update(msg) → new Model + Cmd
  → View() → string (terminal output)
```

Each screen (dashboard, feed list, reader) is a sub-model with its own `Update` and `View`. The root model routes messages to the active screen.

Why bubbletea specifically:
- Part of the Charm ecosystem (lipgloss for styling, glamour for markdown) — designed to work together
- Well-documented with many examples
- Active community and maintenance
- The Elm pattern maps naturally to Go's sequential style — no async/await complexity

## Consequences

**Easier:**
- State is always in one place — no hunting for where a value changed
- Testing: `Update` is a pure function, feed it a message, assert the new state
- Adding new screens: just add a new sub-model, wire it into the router
- Background work is clean — `tea.Cmd` returns messages, no shared state

**Harder:**
- Deeply nested state updates are verbose in Go (no spread operator or immutable data structures)
- Every interaction is a message — simple things like "toggle a boolean" need a message type
- Learning curve if unfamiliar with the Elm pattern

**Giving up:**
- Direct mutation — can't just set a field from a callback, must go through the message loop
- Widget library — bubbletea is lower-level than tview, more assembly required for tables/lists (mitigated by bubbles, Charm's component library)
