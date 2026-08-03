# TOML Storage + Arrow-Key Picker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Switch `favs`'s storage format from JSON to TOML, and replace the picker's numbered-entry selection with arrow-key navigation plus fzf-style type-ahead filtering.

**Architecture:** `internal/store` swaps its JSON encode/decode calls for `BurntSushi/toml`, keeping the same public `Store` API. `internal/picker` gets a new pure `list.Item` adapter (`commandItem`) and its `Pick()` function is rewritten around a `charmbracelet/bubbletea` program driving a `charmbracelet/bubbles/list.Model`, still scoped to `/dev/tty` so stdout stays clean for the shell widget.

**Tech Stack:** Go, Cobra (existing), `github.com/BurntSushi/toml`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles/list`.

## Global Constraints

- Storage file moves to `~/.config/favs/commands.toml`. Clean cutover — no migration code for the old `commands.json`.
- TOML shape is array-of-tables under the key `command` (i.e. `[[command]]` blocks).
- New picker: typing filters immediately (no `/` needed first), arrows/`j`/`k` move the highlight, `Enter` selects, `Esc`/`Ctrl+C` quit with no selection.
- `favs list` (non-interactive) is unaffected — keeps using `picker.FormatList` to stdout.
- The picker must render only to `/dev/tty`, never to real stdout (the shell widget's `$(favs)` depends on this).

---

### Task 1: Storage — JSON to TOML

**Files:**
- Modify: `internal/store/store.go`
- Modify: `internal/store/store_test.go`
- Modify: `go.mod`, `go.sum` (via `go get`)
- Modify: `README.md` (Storage section)

**Interfaces:**
- Produces: `store.Command` (same fields: `ID int`, `Command string`, `Tags []string`, `Description string`, now with `toml:"..."` tags instead of `json:"..."`). `store.Store`, `store.New()`, `store.NewAt(path string)`, `(*Store).Load() ([]Command, error)`, `(*Store).Save([]Command) error`, `(*Store).Add(...)`, `(*Store).Remove(id int) error`, `(*Store).Filter(tag string) ([]Command, error)` — all signatures unchanged from current code.

- [ ] **Step 1: Write the failing test**

In `internal/store/store_test.go`, change the `newTestStore` helper's filename and add a new test asserting the on-disk format is TOML:

```go
func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	return store.NewAt(filepath.Join(t.TempDir(), "commands.toml"))
}
```

Add this test (needs `"os"` and `"strings"` added to the import block):

```go
func TestSave_WritesTOMLFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.toml")
	s := store.NewAt(path)
	if _, err := s.Add("echo hello", []string{"foo", "bar"}, "test desc"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[[command]]") {
		t.Errorf("expected TOML array-of-tables syntax, got:\n%s", content)
	}
	if !strings.Contains(content, `command = "echo hello"`) {
		t.Errorf("expected command field, got:\n%s", content)
	}
	if !strings.Contains(content, "tags = [") {
		t.Errorf("expected tags array, got:\n%s", content)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/store/... -run TestSave_WritesTOMLFormat -v`
Expected: FAIL — current implementation writes JSON, so the file won't contain `[[command]]`.

- [ ] **Step 3: Add the TOML dependency**

Run: `go get github.com/BurntSushi/toml@latest`

This updates `go.mod`/`go.sum` (and may bump the `go` directive — that's expected).

- [ ] **Step 4: Rewrite `internal/store/store.go`**

```go
package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Command struct {
	ID          int      `toml:"id"`
	Command     string   `toml:"command"`
	Tags        []string `toml:"tags"`
	Description string   `toml:"description"`
}

type document struct {
	Command []Command `toml:"command"`
}

type Store struct {
	path string
}

func New() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "favs", "commands.toml")}, nil
}

func NewAt(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() ([]Command, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Command{}, nil
	}
	if err != nil {
		return nil, err
	}
	var doc document
	if _, err := toml.Decode(string(data), &doc); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", s.path, err)
	}
	if doc.Command == nil {
		return []Command{}, nil
	}
	return doc.Command, nil
}

func (s *Store) Save(commands []Command) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(document{Command: commands})
}

func (s *Store) Add(command string, tags []string, description string) (Command, error) {
	commands, err := s.Load()
	if err != nil {
		return Command{}, err
	}
	cmd := Command{
		ID:          nextID(commands),
		Command:     command,
		Tags:        tags,
		Description: description,
	}
	return cmd, s.Save(append(commands, cmd))
}

func (s *Store) Remove(id int) error {
	commands, err := s.Load()
	if err != nil {
		return err
	}
	for i, cmd := range commands {
		if cmd.ID == id {
			return s.Save(append(commands[:i], commands[i+1:]...))
		}
	}
	return fmt.Errorf("no command with id %d", id)
}

func (s *Store) Filter(tag string) ([]Command, error) {
	commands, err := s.Load()
	if err != nil {
		return nil, err
	}
	if tag == "" {
		return commands, nil
	}
	var filtered []Command
	for _, cmd := range commands {
		for _, t := range cmd.Tags {
			if t == tag {
				filtered = append(filtered, cmd)
				break
			}
		}
	}
	return filtered, nil
}

func nextID(commands []Command) int {
	max := 0
	for _, cmd := range commands {
		if cmd.ID > max {
			max = cmd.ID
		}
	}
	return max + 1
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/store/... -v`
Expected: PASS — all existing store tests plus `TestSave_WritesTOMLFormat`.

- [ ] **Step 6: Update README storage section**

In `README.md`, change:

```markdown
## Storage

Commands are stored in `~/.config/favs/commands.json`. Edit it directly if you like.
```

to:

```markdown
## Storage

Commands are stored in `~/.config/favs/commands.toml`. Edit it directly if you like.
```

- [ ] **Step 7: Commit**

```bash
git add internal/store/store.go internal/store/store_test.go go.mod go.sum README.md
git commit -m "feat: switch storage format from JSON to TOML"
```

---

### Task 2: Picker item adapter (`list.Item`)

**Files:**
- Create: `internal/picker/item.go`
- Create: `internal/picker/item_test.go`

**Interfaces:**
- Consumes: `store.Command` (from Task 1 — unchanged shape: `Command`, `Tags []string`, `Description string`).
- Produces: `commandItem` (unexported struct), `newCommandItem(c store.Command) commandItem`, and methods `Title() string`, `Description() string`, `FilterValue() string`, `Command() string`. Task 3 relies on `newCommandItem` and the `Command()` accessor.

- [ ] **Step 1: Write the failing tests**

Create `internal/picker/item_test.go` (white-box test, same package as the implementation so it can reach the unexported type):

```go
package picker

import (
	"strings"
	"testing"

	"github.com/wstevenson1/favs/internal/store"
)

func TestCommandItem_TitleWithTags(t *testing.T) {
	item := newCommandItem(store.Command{Command: "git log --oneline", Tags: []string{"git", "log"}})
	want := "[git,log] git log --oneline"
	if got := item.Title(); got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestCommandItem_TitleWithoutTags(t *testing.T) {
	item := newCommandItem(store.Command{Command: "df -h"})
	want := "df -h"
	if got := item.Title(); got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestCommandItem_Description(t *testing.T) {
	item := newCommandItem(store.Command{Command: "df -h", Description: "Disk usage"})
	if got := item.Description(); got != "Disk usage" {
		t.Errorf("Description() = %q, want %q", got, "Disk usage")
	}
}

func TestCommandItem_FilterValueIncludesTagsAndDescription(t *testing.T) {
	item := newCommandItem(store.Command{
		Command:     "kubectl get pods",
		Tags:        []string{"k8s"},
		Description: "List pods",
	})
	fv := item.FilterValue()
	for _, want := range []string{"kubectl get pods", "k8s", "List pods"} {
		if !strings.Contains(fv, want) {
			t.Errorf("FilterValue() = %q, missing %q", fv, want)
		}
	}
}

func TestCommandItem_Command(t *testing.T) {
	item := newCommandItem(store.Command{Command: "echo hi"})
	if got := item.Command(); got != "echo hi" {
		t.Errorf("Command() = %q, want %q", got, "echo hi")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/picker/... -run TestCommandItem -v`
Expected: FAIL with build errors — `newCommandItem`/`commandItem` don't exist yet.

- [ ] **Step 3: Create `internal/picker/item.go`**

```go
package picker

import (
	"fmt"
	"strings"

	"github.com/wstevenson1/favs/internal/store"
)

type commandItem struct {
	cmd store.Command
}

func newCommandItem(c store.Command) commandItem {
	return commandItem{cmd: c}
}

func (i commandItem) Title() string {
	if len(i.cmd.Tags) > 0 {
		return fmt.Sprintf("[%s] %s", strings.Join(i.cmd.Tags, ","), i.cmd.Command)
	}
	return i.cmd.Command
}

func (i commandItem) Description() string {
	return i.cmd.Description
}

func (i commandItem) FilterValue() string {
	return strings.Join(append([]string{i.cmd.Command, i.cmd.Description}, i.cmd.Tags...), " ")
}

func (i commandItem) Command() string {
	return i.cmd.Command
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/picker/... -v`
Expected: PASS — the new `TestCommandItem_*` tests, plus the still-present `select_test.go` tests (untouched until Task 3).

- [ ] **Step 5: Commit**

```bash
git add internal/picker/item.go internal/picker/item_test.go
git commit -m "feat: add list.Item adapter for picker commands"
```

---

### Task 3: Arrow-key + type-ahead picker

**Files:**
- Modify: `internal/picker/picker.go`
- Delete: `internal/picker/select.go`
- Delete: `internal/picker/select_test.go`
- Create: `internal/picker/picker_test.go`
- Modify: `go.mod`, `go.sum` (via `go get`)
- Modify: `README.md` (Picker section)

**Interfaces:**
- Consumes: `newCommandItem`, `commandItem.Command()` (Task 2); `store.Command` (Task 1).
- Produces: `picker.Pick(commands []store.Command) (string, error)` — same signature as before, still used by `cmd/root.go`. No other package touches `picker.ParseSelection`/`ErrQuit`/`ErrInvalid`, so deleting them is safe (verified by `grep -rn "ParseSelection\|picker.ErrQuit\|picker.ErrInvalid" cmd/` returning nothing before this task starts).

Note on testing: the interactive event loop isn't practically unit-testable outside a real terminal (per the design spec). This task verifies the empty-list short-circuit (the one path that doesn't touch `/dev/tty`) with an automated test, and everything else via `go build`/`go vet` plus a manual smoke test.

- [ ] **Step 1: Confirm nothing outside `internal/picker` uses the numbered-entry API**

Run: `grep -rn "ParseSelection\|picker\.ErrQuit\|picker\.ErrInvalid" --include=*.go . | grep -v internal/picker`
Expected: no output.

- [ ] **Step 2: Delete the numbered-entry files**

```bash
git rm internal/picker/select.go internal/picker/select_test.go
```

- [ ] **Step 3: Add the bubbletea/bubbles dependencies**

Run: `go get github.com/charmbracelet/bubbletea@latest github.com/charmbracelet/bubbles@latest`

- [ ] **Step 4: Rewrite `internal/picker/picker.go`**

```go
package picker

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/wstevenson1/favs/internal/store"
)

type pickerModel struct {
	list     list.Model
	selected string
}

func newPickerModel(commands []store.Command) pickerModel {
	items := make([]list.Item, len(commands))
	for i, c := range commands {
		items[i] = newCommandItem(c)
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "favs"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return pickerModel{list: l}
}

// Init sends a synthetic "/" keypress so the list starts already in
// filtering mode: the user can type immediately (fzf-style) instead of
// pressing "/" first.
func (m pickerModel) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	}
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		// Ctrl+C must always quit, even mid-filter, where bubbles/list
		// would otherwise swallow it as ordinary filter input.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			// No item is selected when the filter matches nothing —
			// ignore Enter instead of quitting with an empty selection.
			if item, ok := m.list.SelectedItem().(commandItem); ok {
				m.selected = item.Command()
				return m, tea.Quit
			}
			return m, nil
		}
		if m.list.FilterState() != list.Filtering && (msg.String() == "esc" || msg.String() == "q") {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	return m.list.View()
}

func Pick(commands []store.Command) (string, error) {
	if len(commands) == 0 {
		fmt.Fprintln(os.Stderr, "no saved commands. use 'favs add' to get started")
		return "", nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("could not open terminal: %w", err)
	}
	defer tty.Close()

	p := tea.NewProgram(newPickerModel(commands), tea.WithInput(tty), tea.WithOutput(tty))
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("picker error: %w", err)
	}
	return finalModel.(pickerModel).selected, nil
}
```

- [ ] **Step 5: Create `internal/picker/picker_test.go`**

```go
package picker_test

import (
	"testing"

	"github.com/wstevenson1/favs/internal/picker"
)

func TestPick_EmptyListReturnsEmptyStringWithoutError(t *testing.T) {
	selected, err := picker.Pick(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "" {
		t.Errorf("expected empty selection, got %q", selected)
	}
}
```

This works without a real `/dev/tty` because the empty-list check returns before the picker ever opens the terminal.

- [ ] **Step 6: Build and test**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: build succeeds, vet is clean, all tests pass (store tests, item tests, the new empty-list picker test).

- [ ] **Step 7: Manual smoke test**

```bash
go build -o /tmp/favs-test .
/tmp/favs-test add "echo one" --tags demo --desc "first"
/tmp/favs-test add "echo two" --tags demo --desc "second"
/tmp/favs-test
```

Confirm: the list appears already filterable (typing narrows it), Up/Down moves the highlight, Enter prints the selected command to stdout and exits, Esc (after clearing/leaving the filter) and Ctrl+C both quit with no output.

- [ ] **Step 8: Update README picker section**

In `README.md`, replace:

```markdown
## Picker

```
  1  [ssh]   ssh aeneas64-ubuntu  aeneas64 Ubuntu VM
  2  [git]   git log --oneline --graph --all  Pretty git log
  3  [sys]   df -h  Disk usage human readable

Select (1-3, q to quit): _
```

Select a number and the command is placed in your readline buffer ready to edit — `{placeholder}` values and all.
```

with:

```markdown
## Picker

```
favs
```

Type to filter (no need to press `/` first), use the arrow keys (or `j`/`k`) to move the highlight, and press Enter to select. The chosen command is placed in your readline buffer ready to edit — `{placeholder}` values and all. `Esc` or `Ctrl+C` quits without selecting anything.
```

- [ ] **Step 9: Commit**

```bash
git add internal/picker/picker.go internal/picker/picker_test.go go.mod go.sum README.md
git commit -m "feat: replace numbered-entry picker with arrow-key + type-ahead filter"
```
