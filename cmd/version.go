package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stacksenv/cli/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("StacksENV v" + version.Version + "/" + version.CommitSHA)
	},
}
