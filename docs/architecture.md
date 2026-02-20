# Architecture

## Overview

Feeder follows the Elm architecture via bubbletea: a single model owns all state, messages drive updates, and the view is a pure function of state. Background work (feed fetching, article extraction) runs in goroutines and communicates back through bubbletea's command system.

## Data Flow

```mermaid
graph TD
    Config[config.toml] -->|read on startup| App[App Model]

    App -->|tea.Cmd| Fetcher[Feed Fetcher]
    Fetcher -->|goroutines + errgroup| RSS[gofeed - RSS/Atom]
    Fetcher -->|goroutines + errgroup| GH[go-github - Releases]

    RSS -->|FeedUpdatedMsg| App
    GH -->|FeedUpdatedMsg| App
    Fetcher -->|FeedErrorMsg| App

    App -->|dedup + write| Cache[cache.json]
    Cache -->|read| App

    App -->|on mark read| State[state.json]
    State -->|read on startup| App

    App -->|on bookmark| Bookmarks[bookmarks.md]

    App -->|on open article| Extractor[go-readability]
    Extractor -->|ArticleExtractedMsg| App

    App -->|View| TUI[Terminal UI]
    TUI -->|KeyMsg| App
```

## Screen Flow

```mermaid
stateDiagram-v2
    [*] --> Dashboard
    Dashboard --> FeedList: Enter (select feed/tag)
    Dashboard --> Dashboard: r (refresh), a (all articles)
    FeedList --> Dashboard: Esc / d
    FeedList --> Reader: Enter (open article)
    FeedList --> FeedList: m (toggle read), b (bookmark)
    Reader --> FeedList: Esc
    Reader --> Dashboard: d
    Reader --> Reader: j/k (scroll), o (browser), b (bookmark), y (yank)
    Dashboard --> [*]: q
```

## Storage Layout

```
~/.config/feeder/
└── config.toml              # Feed list + settings (source of truth)

~/.cache/feeder/
└── cache.json               # Fetched articles (ephemeral, local only)

~/Documents/ (or configured path)
├── feeder-state.json        # Read article GUIDs (synced via GitHub)
└── feeder-bookmarks.md      # Saved articles (synced via GitHub)
```

## Concurrency Model

```
Main goroutine (bubbletea)
│
├── Ticker goroutine ──── sends TickMsg every N minutes
│
└── Refresh command ──── spawns errgroup
    ├── fetch feed 1 ─┐
    ├── fetch feed 2 ──┤  semaphore limits concurrency
    ├── fetch feed 3 ──┤
    └── fetch feed N ─┘
         │
         └── each sends FeedUpdatedMsg or FeedErrorMsg
```

No shared mutable state. The bubbletea model is only touched by the main goroutine. Background goroutines communicate exclusively through messages.
