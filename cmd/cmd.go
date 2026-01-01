package cmd

import (
	"os"
	"strings"
)

// Execute executes the commands.
func Execute() error {
	// Disable flag parsing if args should be passed to system commands
	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// List of known stacksenv commands
		knownCommands := []string{"set", "init", "update", "remote", "version"}

		// If first arg starts with stacksenv://, disable flag parsing
		if strings.HasPrefix(firstArg, "stacksenv://") {
			rootCmd.DisableFlagParsing = true
		} else {
			// Check if first arg is a known stacksenv command
			isKnownCommand := false
			for _, cmd := range knownCommands {
				if firstArg == cmd {
					isKnownCommand = true
					break
				}
			}

			// If it's not a known command, disable flag parsing to pass args to system commands
			if !isKnownCommand && !strings.HasPrefix(firstArg, "-") {
				rootCmd.DisableFlagParsing = true
			}
		}
	}

	return rootCmd.Execute()
}
