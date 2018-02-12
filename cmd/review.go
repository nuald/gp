package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reviewCmd)
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Add #review hashtag and the list of reviewers into the HEAD commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdArgs := []string{"show", "-s", "--format=%B"}
		/* #nosec */
		out, err := exec.Command("git", cmdArgs...).Output()
		if err != nil {
			return errors.Wrap(err, 1)
		}

		comment := string(out)
		if strings.Contains(comment, "#review") {
			return nil
		}

		reviewers, err := trim(readConfig("reviewers", "Reviewers", false, false))
		if err != nil {
			return err
		}

		msg := fmt.Sprintf(`-m%s

#review
%s`, comment, reviewers)
		cmdArgs = []string{"commit", "--amend", msg}
		/* #nosec */
		if err := exec.Command("git", cmdArgs...).Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}
