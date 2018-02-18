package cmd

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var addReview bool

func init() {
	rootCmd.AddCommand(shelveCmd)
	shelveCmd.PersistentFlags().BoolVarP(&addReview, "add-review",
		"a", true, "add review comment")
}

var shelveCmd = &cobra.Command{
	Use:   "shelve",
	Short: "Shelve changes back to the p4 repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := prepareSubmit()
		if err != nil {
			return err
		}

		if addReview {
			// Add initial #review if needed
			if err = updateReviewHashtag(); err != nil {
				return err
			}
		}

		gitCmd := newCmd("git", "p4", "submit", "--shelve")
		if err = gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		cl, err := getPendingCL(workspace)
		if err != nil {
			return err
		}

		commits, err := getCommits()
		if err != nil {
			return err
		}

		for index, sha := range commits {
			if err = addNote(cl[index], sha); err != nil {
				return err
			}
		}

		if addReview {
			if err = writePendingChanges(cl[:len(commits)]); err != nil {
				return err
			}

			// Update #review with the number of Swarm review
			if err = updateReviewHashtag(); err != nil {
				return err
			}
		}

		return nil
	},
}

func updateReviewHashtag() error {
	_, err := getReviewers(reviewersGroup)
	if err != nil {
		return err
	}

	reviewCmd := "gp review -r " + reviewersGroup
	args := []string{"rebase", "-x", reviewCmd, "p4/HEAD"}
	cmd := newCmd("git", args...)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}

func getPendingCL(workspace string) ([]string, error) {
	args := []string{"changes", "-s", "shelved", "-c", workspace}
	/* #nosec */
	out, err := exec.Command("p4", args...).Output()
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	var result []string
	clRe := regexp.MustCompile(`^Change (\d+) on .*$`)
	for _, line := range strings.Split(string(out), "\n") {
		m := clRe.FindStringSubmatch(line)
		if m != nil {
			result = append(result, m[1])
		}
	}

	return result, nil
}

func getCommits() ([]string, error) {
	args := []string{"rev-list", "--no-merges", "p4/HEAD.."}
	/* #nosec */
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}

	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}
