package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const shellWidget = `
_favs_widget() {
  local selected
  selected=$(favs 2>/dev/null)
  if [[ -n "$selected" ]]; then
    READLINE_LINE="$selected"
    READLINE_POINT=${#READLINE_LINE}
  fi
}
bind -x '"\C-f": _favs_widget'
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Print bash shell widget for readline integration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(shellWidget)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
