package stackenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func HandleStacksenvURLCLI(url string, args []string) error {
	fmt.Println("Handling stacksenv URL: ", args, url)

	if url != "" {
		config, err := ParseURL(url)
		if err != nil {
			return fmt.Errorf("failed to parse stacksenv URL: %w", err)
		}
		fmt.Println("Handling stacksenv URL: ", config)
	}

	// Execute remaining args as system CLI commands (e.g., "node -v", "python -v")
	if len(args) > 0 {
		// Parse the command and its arguments
		command := args[0]
		commandArgs := args[1:]

		// Execute the system command
		cmd := exec.Command(command, commandArgs...)

		// Set up output to stdout and stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		// Execute the command
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute command '%s %s': %w", command, strings.Join(commandArgs, " "), err)
		}
	}

	return nil
}
