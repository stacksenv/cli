package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().String("serverurl", "", "Set the server URL in the global configuration")
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a value for a key",
	Long:  `Set a value for a key in the global configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, err := cmd.Flags().GetString("serverurl")
		if err != nil {
			return err
		}

		// If serverurl flag is provided, update the global config
		if serverURL != "" {
			if err := updateGlobalConfig("serverurl", serverURL); err != nil {
				return err
			}
			fmt.Printf("Successfully updated serverurl to: %s\n", serverURL)
			return nil
		}

		fmt.Println("set called with args:", args)
		return nil
	},
}
