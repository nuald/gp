package cmd

import (
	"github.com/spf13/cobra"
	"github.com/go-errors/errors"
)

func init() {
	rootCmd.AddCommand(cloneCmd)
}

var cloneCmd = &cobra.Command{
	Use:   "clone <repository> [<directory>]",
	Short: "Create a new Git directory from an existing p4 repository",
	Long: `Create a new Git directory from an existing p4 repository specified
by the depot and the project (or the stream) paths:

    gp clone //depot/project
    gp clone //depot/stream destination

To reproduce the entire p4 history in Git, please use the @all modifier
on the depot path:

    gp clone //depot/project@all
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := login(); err != nil {
			return err
		}

		cmdArgs := append([]string{"p4", "clone"}, args...)
		gitCmd := newCmd("git", cmdArgs...)
		if err := gitCmd.Run(); err != nil {
			return errors.Wrap(err, 1)
		}

		return nil
	},
}
