# Package: `internal/ui`

> Fill in as Phase 3 is built.

## Purpose

TUI screens built with bubbletea. Each screen is a sub-model with its own `Update` and `View` methods.

## Key Files

| File | Responsibility |
|------|---------------|
| `dashboard.go` | Feed table with unread counts, tag filtering |
| `feedlist.go` | Article list for a selected feed or tag |
| `reader.go` | Full article view with glamour markdown rendering |
| `help.go` | Keybinding overlay |

## Concepts to Document

- [ ] Bubbletea sub-model pattern — how the root model delegates to screens
- [ ] lipgloss styling — colors, borders, layout
- [ ] glamour markdown rendering — configuration, custom styles
- [ ] Viewport scrolling for the reader
- [ ] Key handling — how keys route through the model hierarchy
- [ ] Responsive layout — adapting to terminal size via `tea.WindowSizeMsg`
