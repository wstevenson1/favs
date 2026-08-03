# favs — TOML storage + arrow-key picker — Design Spec

**Date:** 2026-08-03
**Status:** Approved

## Overview

Two changes to the `favs` CLI ([base design](2026-06-07-favs-design.md)):

1. Storage format changes from JSON to TOML.
2. The interactive picker changes from numbered-entry (type a number, press Enter) to arrow-key navigation with fzf-style type-ahead filtering.

---

## Storage — JSON → TOML

**File:** `~/.config/favs/commands.toml` (was `commands.json`). Clean cutover — no migration code; there's no meaningful existing user data.

**Library:** `github.com/BurntSushi/toml`.

**Shape** — array-of-tables:

```toml
[[command]]
id = 1
command = "kubectl get pods -n {namespace}"
tags = ["k8s", "kubectl"]
description = "List pods in a namespace"
```

`store.Command` gets `toml:"..."` struct tags (replacing `json:"..."`). Because BurntSushi's encoder needs a named top-level key to emit an array-of-tables, `Load`/`Save` wrap the slice in an unexported struct:

```go
type document struct {
    Command []Command `toml:"command"`
}
```

`Load` uses `toml.Decode`, `Save` uses `toml.NewEncoder(file).Encode`. Behavior is otherwise unchanged: missing file → empty slice, corrupted file → error (message updates from "error reading ... json" to "... toml").

`store` package tests are updated to use `.toml` fixtures.

---

## Picker — numbered entry → arrow nav + type-ahead filter

**New dependencies:** `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles/list` (pulls in `lipgloss` and `sahilm/fuzzy` transitively).

`internal/picker/picker.go`'s `Pick()` is rewritten around a `tea.Program`:

- Still opens `/dev/tty` directly, passed as both `tea.WithInput(tty)` and `tea.WithOutput(tty)`. The interactive UI renders only to the tty — real stdout stays clean so the shell widget's `$(favs)` continues to capture only the final selected command, unchanged from the base design.
- Each `store.Command` is adapted to `list.Item` (`Title()`/`Description()`/`FilterValue()`), reusing the tags/description display convention from `FormatList`.
- Typing immediately filters the list (bubbles' built-in fuzzy filtering) — no `/` key needed first, matching fzf muscle memory. Backspace edits the filter. Up/Down (or j/k) moves the highlight.
- `Enter` selects the highlighted item and returns its `Command` string. `Esc`/`Ctrl+C` quits with no selection — same contract as today's `q`.
- `internal/picker/select.go` (`ParseSelection`, `ErrQuit`, `ErrInvalid`) and its tests are deleted; numbered-entry parsing no longer applies.

`favs list` (non-interactive, scriptable output) is untouched — still plain text via `FormatList` to stdout.

**Testing:** the picker is now a real TUI event loop and isn't practically unit-testable the way numbered-entry parsing was. We unit-test the `list.Item` adapter (title/description/filter-value formatting) as a pure function; the interactive loop itself is verified manually, consistent with the base spec's "no integration tests, surface area is small" approach.

---

## Error Handling (deltas from base spec)

| Scenario | Behavior |
|----------|----------|
| TOML file missing | Create with empty document on first write (unchanged behavior, new format) |
| TOML file corrupted | Print `"error reading ~/.config/favs/commands.toml: <err>"` to stderr, exit 1 |
| Picker invoked with empty list | Unchanged: print `"no saved commands. use 'favs add' to get started"`, exit 0 |
| Picker: no filter matches | List shows empty; Enter is a no-op until filter is edited or cleared |

---

## Out of Scope

- Migrating existing `commands.json` data — clean cutover only.
- Config options for picker theming/colors — use bubbles/lipgloss defaults.
