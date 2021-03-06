package cmd

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(submitCmd)
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit changes back to the p4 repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := prepareSubmit(); err != nil {
			return err
		}

		gitCmd := newCmd("git", "p4", "submit")
		if err := gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}
