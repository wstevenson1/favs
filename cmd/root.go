package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var tagFilter string

var rootCmd = &cobra.Command{
	Use:   "favs",
	Short: "Bookmark and recall favorite shell commands",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&tagFilter, "tag", "", "filter by tag")
}
