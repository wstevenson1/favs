package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wstevenson1/favs/internal/picker"
	"github.com/wstevenson1/favs/internal/store"
)

var tagFilter string

var rootCmd = &cobra.Command{
	Use:   "favs",
	Short: "Bookmark and recall favorite shell commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New()
		if err != nil {
			return err
		}
		commands, err := s.Filter(tagFilter)
		if err != nil {
			return err
		}
		selected, err := picker.Pick(commands)
		if err != nil {
			return err
		}
		if selected != "" {
			fmt.Println(selected)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&tagFilter, "tag", "", "filter by tag")
}
