package cmd

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rebaseCmd)
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Update the Git repository with recent changes from p4",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := login(); err != nil {
			return err
		}

		gitCmd := newCmd("git", "p4", "rebase")
		if err := gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}
