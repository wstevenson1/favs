package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wstevenson1/favs/internal/store"
)

var addTags string
var addDesc string

var addCmd = &cobra.Command{
	Use:   "add <command>",
	Short: "Add a command to favorites",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}
		var tags []string
		if addTags != "" {
			tags = strings.Split(addTags, ",")
		}
		added, err := s.Add(args[0], tags, addDesc)
		if err != nil {
			return err
		}
		fmt.Printf("Added command #%d\n", added.ID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addTags, "tags", "", "comma-separated tags")
	addCmd.Flags().StringVar(&addDesc, "desc", "", "short description")
	rootCmd.AddCommand(addCmd)
}
