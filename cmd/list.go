package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wstevenson1/favs/internal/picker"
	"github.com/wstevenson1/favs/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}
		commands, err := s.Filter(tagFilter)
		if err != nil {
			return err
		}
		if len(commands) == 0 {
			fmt.Println("no saved commands. use 'favs add' to get started")
			return nil
		}
		fmt.Print(picker.FormatList(commands))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
