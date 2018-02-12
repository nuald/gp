package cmd

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var addReview bool

func init() {
	rootCmd.AddCommand(shelveCmd)
	shelveCmd.PersistentFlags().BoolVarP(&addReview, "add-review",
		"r", true, "add review comment")
}

var shelveCmd = &cobra.Command{
	Use:   "shelve",
	Short: "Shelve changes back to the p4 repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := prepareSubmit(); err != nil {
			return err
		}

		if addReview {
			if err := addReviewHashtag(); err != nil {
				return err
			}
		}

		gitCmd := newCmd("git", "p4", "submit", "--shelve")
		if err := gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}

func addReviewHashtag() error {
	_, err := trim(readConfig("reviewers", "Reviewers", false, false))
	if err != nil {
		return err
	}

	args := []string{"rebase", "-x", "gp review", "p4/HEAD"}
	cmd := newCmd("git", args...)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}
