package cmd

import (
	"log"

	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		err := login()
		if err != nil {
			log.Fatal(err)
			return
		}

		cmdArgs := append([]string{"p4", "clone"}, args...)
		gitCmd := newCmd("git", cmdArgs...)
		gitCmd.Run()
	},
}
