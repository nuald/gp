package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shelveCmd)
}

var shelveCmd = &cobra.Command{
	Use:   "shelve",
	Short: "Shelve changes back to the p4 repository",
	Run: func(cmd *cobra.Command, args []string) {
		err := login()
		if err != nil {
			log.Fatal(err)
			return
		}

		gitCmd := newCmd("git", "p4", "submit", "--shelve")
		gitCmd.Run()
	},
}
