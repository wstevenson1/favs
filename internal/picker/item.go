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
