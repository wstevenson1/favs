package picker

import (
	"fmt"
	"strings"

	"github.com/wstevenson1/favs/internal/store"
)

func FormatList(commands []store.Command) string {
	var sb strings.Builder
	for i, cmd := range commands {
		tags := ""
		if len(cmd.Tags) > 0 {
			tags = fmt.Sprintf("[%s]", strings.Join(cmd.Tags, ","))
		}
		desc := ""
		if cmd.Description != "" {
			desc = "  " + cmd.Description
		}
		if tags != "" {
			fmt.Fprintf(&sb, "%3d  %-16s %s%s\n", i+1, tags, cmd.Command, desc)
		} else {
			fmt.Fprintf(&sb, "%3d  %s%s\n", i+1, cmd.Command, desc)
		}
	}
	return sb.String()
}
