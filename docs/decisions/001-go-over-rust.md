# 001 — Go over Rust

## Status

Accepted

## Context

The original plan used Rust for its performance and single-binary output. However, Feeder is a personal TUI RSS reader — not a high-performance system. The workload is: fetch a few dozen feeds over HTTP, parse XML/JSON, render text in a terminal, and read/write small files.

Rust's strengths (memory safety without GC, zero-cost abstractions) aren't bottlenecks here. Its costs are real though: slow compile times, borrow checker friction for a stateful TUI app, and a larger dependency tree (tokio, reqwest, feed-rs, rusqlite, etc.) for things Go's stdlib handles natively.

## Decision

Use Go. It provides:

- **Single binary** — same as Rust, no runtime to install
- **Fast compilation** — seconds, not minutes
- **Built-in concurrency** — goroutines replace tokio entirely
- **Rich stdlib** — `net/http`, `encoding/json`, `time` cover what Rust needs 5+ crates for
- **Mature TUI ecosystem** — Charm's bubbletea uses the same Elm pattern we wanted
- **Simpler learning curve** — fewer concepts to fight while learning the domain (feeds, TUI, file sync)

## Consequences

**Easier:**
- Faster iteration cycles (compile + run in seconds)
- Concurrency is trivial (goroutines + channels vs. async/await + pinning)
- Fewer dependencies to manage
- Cross-compilation is a single `GOOS`/`GOARCH` env var

**Harder:**
- No sum types — screen/message enums require interfaces + type switches instead of exhaustive pattern matching
- Error handling is verbose (`if err != nil` everywhere)
- No ownership model — need discipline around not sharing mutable state across goroutines

**Giving up:**
- Rust learning (was a secondary goal, not the primary one)
- Marginal performance (irrelevant at this scale)
