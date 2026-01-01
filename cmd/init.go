package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new project",
	Long:  `Initialize a new project by creating a .stacksenv/config.json file in the current directory.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := createLocalConfig(); err != nil {
			return err
		}

		cwd, _ := os.Getwd()
		configPath := filepath.Join(cwd, ".stacksenv", "config.json")
		fmt.Printf("Initialized project configuration at: %s\n", configPath)
		return nil
	},
}
