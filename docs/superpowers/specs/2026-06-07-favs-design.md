# favs — Design Spec

**Date:** 2026-06-07  
**Status:** Approved

## Overview

A Go CLI tool (`favs`) for bookmarking favorite shell commands. Commands are stored with optional tags and descriptions. An interactive picker lets you select a command and inject it into the bash readline buffer for editing — without executing it. Similar in spirit to navi's snippets feature.

---

## Architecture & Components

Two artifacts:

1. **`favs` binary** — Go, using Cobra for subcommands. Handles storage, CRUD operations, and the interactive picker.
2. **Bash shell widget** — a small function sourced in `.bashrc` that invokes the binary, captures its stdout, and injects the result into the readline buffer via `READLINE_LINE`.

### Internal packages

| Package | Responsibility |
|---------|---------------|
| `store` | Reads/writes the JSON file; owns the data model and ID assignment |
| `picker` | Renders the numbered list to `/dev/tty`, reads selection from `/dev/tty`, returns the selected command string to stdout |
| `cmd` | Cobra subcommand wiring |

The picker uses `/dev/tty` for both display and input so that the shell widget's `$()` subshell only captures the selected command string — not the picker UI.

---

## Data Model & Storage

**File:** `~/.config/favs/commands.json`

```json
[
  {
    "id": 1,
    "command": "kubectl get pods -n {namespace}",
    "tags": ["k8s", "kubectl"],
    "description": "List pods in a namespace"
  }
]
```

| Field | Type | Notes |
|-------|------|-------|
| `id` | int | Stable, monotonically incrementing, never reused. `max(existing) + 1` on add. |
| `command` | string | Raw command string. `{placeholder}` syntax is a convention only — the tool treats it as plain text. |
| `tags` | []string | Optional. Empty array if none given. |
| `description` | string | Optional short label shown in the picker. |

The file is created automatically (with an empty array) on the first write.

---

## Subcommands & UX

```
favs                              # interactive picker (default)
favs add <command> [flags]        # add a command
favs list [--tag <tag>]           # print all commands (non-interactive)
favs rm <id>                      # remove by ID
favs init                         # print bash shell widget snippet to stdout
```

### `favs` (no args) — interactive picker

Displays a numbered list on `/dev/tty` and prompts for a number. Prints the selected command to stdout and exits.

```
  1  [k8s,kubectl]  kubectl get pods -n {namespace}   List pods in a namespace
  2  [git]          git log --oneline --graph --all
  3  [docker]       docker compose up -d

Select (1-3, q to quit): _
```

Accepts `--tag <tag>` to pre-filter the list before displaying.

### `favs add`

```
favs add "git log --oneline --graph --all" --tags git --desc "Pretty git log"
```

- `--tags` — comma-separated tag list (optional)
- `--desc` — short description (optional)
- Prints the assigned ID on success: `Added command #4`

### `favs list`

Same display format as the picker but non-interactive — prints and exits. Accepts `--tag <tag>` to filter.

### `favs rm <id>`

Removes the entry with the given ID. Prints confirmation on success. Exits 1 with a clear message if ID not found.

### `favs init`

Prints the bash shell widget snippet to stdout for easy setup:

```bash
favs init >> ~/.bashrc && source ~/.bashrc
```

---

## Shell Integration

The snippet printed by `favs init`:

```bash
_favs_widget() {
  local selected
  selected=$(favs 2>/dev/null)
  if [[ -n "$selected" ]]; then
    READLINE_LINE="$selected"
    READLINE_POINT=${#READLINE_LINE}
  fi
}
bind -x '"\C-f": _favs_widget'
```

`READLINE_LINE` and `READLINE_POINT` are bash readline variables. Setting them inside a `bind -x` function places text in the input buffer at the cursor position without executing it.

Default key binding: `Ctrl+F` (mnemonic: **f**avs). The user can change the binding in the snippet before appending to `.bashrc`.

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| JSON file missing | Create with empty array on first write |
| `favs rm` with unknown ID | Print `"no command with id <n>"` to stderr, exit 1 |
| Picker invoked with empty list | Print `"no saved commands. use 'favs add' to get started"`, exit 0 |
| Non-numeric picker input | Re-prompt once, then exit cleanly |
| JSON file corrupted | Print `"error reading ~/.config/favs/commands.json: <err>"` to stderr, exit 1 |

---

## Testing

- **`store` package unit tests:** add, list, remove, ID assignment, auto-creation of missing file
- **`picker` selection logic unit tests:** selection parsing is separated from `/dev/tty` rendering so it can be tested without a terminal
- No integration tests — surface area is small and the shell widget is a one-liner
