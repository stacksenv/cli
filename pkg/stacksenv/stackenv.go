package stacksenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func HandleStacksenvURLCLI(url string, args []string) error {
	var properties []ContextData[any]
	originalURL := url
	url = strings.Replace(url, "stacksenv://", "", 1)
	if url != "" {
		config, err := ParseURL(url)
		if err != nil {
			return fmt.Errorf("failed to parse stacksenv URL: %w", err)
		}
		properties, err = GetContextDecryptedData(&config)
		if err != nil {
			return fmt.Errorf("failed to parse HTTP response: %w", err)
		}

		fmt.Println("Properties: ", len(properties))
		for _, contextData := range properties {
			fmt.Println(contextData.Property, "=", "***")
		}
	}

	// Execute remaining args as system CLI commands (e.g., "node -v", "python -v")
	if len(args) > 0 {
		// Parse the command and its arguments
		command := args[0]
		commandArgs := args[1:]

		// Execute the system command
		cmd := exec.Command(command, commandArgs...)

		// Set environment variables from properties for the command execution
		if originalURL != "" && len(properties) > 0 {
			// Start with current environment
			cmd.Env = os.Environ()

			// Add properties as environment variables
			for _, contextData := range properties {
				envVar := fmt.Sprintf("%s=%s", contextData.Property, contextData.Value.(string))
				cmd.Env = append(cmd.Env, envVar)
			}
		}

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
