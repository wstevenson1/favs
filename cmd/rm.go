package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wstevenson1/favs/internal/store"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove a saved command by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("id must be a number")
		}
		s, err := store.New()
		if err != nil {
			return err
		}
		if err := s.Remove(id); err != nil {
			return err
		}
		fmt.Printf("Removed command #%d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
