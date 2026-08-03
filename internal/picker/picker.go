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
