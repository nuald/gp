package cmd

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shelveCmd)
}

var shelveCmd = &cobra.Command{
	Use:   "shelve",
	Short: "Shelve changes back to the p4 repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := prepareSubmit(); err != nil {
			return err
		}

		gitCmd := newCmd("git", "p4", "submit", "--shelve")
		if err := gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}
