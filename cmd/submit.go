package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(submitCmd)
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit changes back to the p4 repository",
	Run: func(cmd *cobra.Command, args []string) {
		err := login()
		if err != nil {
			log.Fatal(err)
			return
		}

		gitCmd := newCmd("git", "p4", "submit")
		gitCmd.Run()
	},
}
