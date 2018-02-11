package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rebaseCmd)
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Update the Git repository with recent changes from p4",
	Run: func(cmd *cobra.Command, args []string) {
		err := login()
		if err != nil {
			log.Fatal(err)
			return
		}

		gitCmd := newCmd("git", "p4", "rebase")
		gitCmd.Run()
	},
}
