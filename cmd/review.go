package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func getReviewers(group string) (string, error) {
	return trim(readConfig("reviewers."+group, "Reviewers", false, false))
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
			return replaceReview(comment)
		}

		reviewers, err := getReviewers(reviewersGroup)
		if err != nil {
			return err
		}

		return amend(fmt.Sprintf(`%s

#review
%s`, comment, reviewers))
	},
}

func replaceReview(comment string) error {
	pending, err := readPendingChanges()
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		// No additional changes are required
		return nil
	}

	note, pending := pending[len(pending)-1], pending[:len(pending)-1]

	cl, err := strconv.Atoi(note)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	// Swarm review has the (CL + 1) number
	review := "#review-" + strconv.Itoa(cl+1)
	reviewRe := regexp.MustCompile(`(#review(-\d+)?)`)
	msg := reviewRe.ReplaceAllString(comment, review)
	if err = amend(msg); err != nil {
		return err
	}

	// Sometimes notes are lost, so need to write them again
	if err = addNote(note, ""); err != nil {
		return err
	}

	return writePendingChanges(pending)
}

func amend(msg string) error {
	cmdArgs := []string{"commit", "--amend", "-m" + msg}
	/* #nosec */
	if err := exec.Command("git", cmdArgs...).Run(); err != nil {
		return errors.Wrap(err, 1)
	}

	return nil
}
