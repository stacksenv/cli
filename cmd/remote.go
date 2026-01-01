package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteAddCmd)
	remoteAddCmd.AddCommand(remoteAddOriginCmd)
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote projects",
	Long:  `Manage remote projects.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return nil
	},
}

var remoteAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a remote project",
	Long:  `Add a remote project.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return nil
	},
}

var remoteAddOriginCmd = &cobra.Command{
	Use:   "origin  <originurl>",
	Short: "Add an origin remote project",
	Long:  `Add an origin remote project.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, _ []string) error {
		return nil
	},
}
