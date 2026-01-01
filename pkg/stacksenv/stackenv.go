package stacksenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Handler handles stacksenv URL CLI operations including fetching context data
// and executing system commands with the appropriate environment variables.
type Handler struct {
	urlParser       URLParser
	clientService   ClientService
	commandExecutor CommandExecutor
}

// NewHandler creates a new Handler with the provided dependencies.
// If nil is passed for any dependency, a default implementation will be used.
func NewHandler(urlParser URLParser, clientService ClientService, commandExecutor CommandExecutor) *Handler {
	h := &Handler{}

	if urlParser == nil {
		h.urlParser = NewURLParser()
	} else {
		h.urlParser = urlParser
	}

	if clientService == nil {
		httpClient := NewHTTPClient()
		crypto := NewCryptoService()
		h.clientService = NewClientService(httpClient, crypto)
	} else {
		h.clientService = clientService
	}

	if commandExecutor == nil {
		h.commandExecutor = NewCommandExecutor()
	} else {
		h.commandExecutor = commandExecutor
	}

	return h
}

// HandleStacksenvURLCLI processes a stacksenv URL and executes the provided command
// with environment variables from the fetched context data.
//
// The process:
//  1. Parses the stacksenv URL (if provided) to extract configuration
//  2. Fetches and decrypts context data from the server
//  3. Sets environment variables from the context data
//  4. Executes the provided command with those environment variables
//
// Parameters:
//   - url: The stacksenv URL (format: stacksenv://ID:SECRET:SECRET_KEY@SERVER_URL/BRANCH)
//   - args: Command and arguments to execute (e.g., ["node", "-v"] or ["python", "script.py"])
//
// Returns an error if URL parsing, data fetching, or command execution fails.
func (h *Handler) HandleStacksenvURLCLI(url string, args []string) error {
	var properties []ContextData[any]
	originalURL := url

	// Parse and process URL if provided
	if url != "" {
		// Remove protocol prefix if present
		url = strings.TrimPrefix(url, "stacksenv://")

		if url != "" {
			// Parse URL to get configuration
			config, err := h.urlParser.ParseURL(url)
			if err != nil {
				return fmt.Errorf("failed to parse stacksenv URL: %w", err)
			}

			// Fetch and decrypt context data
			properties, err = h.clientService.GetContextDecryptedData(&config)
			if err != nil {
				return fmt.Errorf("failed to fetch context data: %w", err)
			}

			// Log properties (masking sensitive values)
			fmt.Printf("Properties: %d\n", len(properties))
			for _, contextData := range properties {
				fmt.Printf("%s = ***\n", contextData.Property)
			}
		}
	}

	// Execute command if provided
	if len(args) == 0 {
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	// Prepare environment variables from properties
	var envVars []string
	if originalURL != "" && len(properties) > 0 {
		envVars = make([]string, 0, len(properties))
		for _, contextData := range properties {
			// Convert value to string (assuming it's already a string or can be converted)
			value, ok := contextData.Value.(string)
			if !ok {
				// Try to convert other types to string
				value = fmt.Sprintf("%v", contextData.Value)
			}
			envVars = append(envVars, fmt.Sprintf("%s=%s", contextData.Property, value))
		}
	}

	// Execute command with environment variables
	return h.commandExecutor.Execute(command, commandArgs, envVars)
}

// DefaultCommandExecutor is the default implementation of CommandExecutor.
type DefaultCommandExecutor struct{}

// NewCommandExecutor creates a new command executor instance.
func NewCommandExecutor() CommandExecutor {
	return &DefaultCommandExecutor{}
}

// Execute runs a system command with the given arguments and environment variables.
//
// It creates a new process with:
//   - The specified command and arguments
//   - The provided environment variables merged with the current environment
//   - Standard input, output, and error streams connected to the parent process
//
// Returns an error if the command execution fails.
func (e *DefaultCommandExecutor) Execute(command string, args []string, env []string) error {
	cmd := exec.Command(command, args...)

	// Set up I/O streams
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set environment variables
	if len(env) > 0 {
		// Start with current environment
		cmd.Env = os.Environ()
		// Append provided environment variables (they will override existing ones)
		cmd.Env = append(cmd.Env, env...)
	}

	// Execute command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command '%s %s': %w", command, strings.Join(args, " "), err)
	}

	return nil
}

// HandleStacksenvURLCLI is a convenience function that uses default implementations.
// It's maintained for backward compatibility.
func HandleStacksenvURLCLI(url string, args []string) error {
	handler := NewHandler(nil, nil, nil)
	return handler.HandleStacksenvURLCLI(url, args)
}

// HandleStacksENV fetches and returns context data based on the provided configuration.
//
// It supports two modes:
//  1. URL mode: If URL is provided, it parses the URL and fetches properties
//  2. Config mode: If URL is empty, it validates required config properties and fetches properties
//  3. If SetOSEnv is true, it will set the environment variables in the OS environment
//
// Required properties for config mode:
//   - ID: Unique identifier for the environment
//   - Secret: Secret key for authentication
//   - SecretKey: Additional secret key for encryption
//   - ServerURL: Server hostname or IP address
//   - Branch: Branch name (e.g., "dev", "prod")
//
// Returns the context data (properties) or an error if URL parsing, validation, or data fetching fails.
func HandleStacksENV(cnf *RequestConfig) ([]ContextData[any], error) {
	// Create default implementations
	httpClient := NewHTTPClient()
	crypto := NewCryptoService()
	clientService := NewClientService(httpClient, crypto)
	urlParser := NewURLParser()

	var config *Config

	// Determine configuration source
	switch {
	case cnf != nil && cnf.URL != "":
		// URL mode: Parse URL to get configuration
		url := strings.TrimPrefix(cnf.URL, "stacksenv://")
		parsedConfig, err := urlParser.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse stacksenv URL: %w", err)
		}
		config = &parsedConfig

	case cnf != nil && cnf.Config != nil:
		// Config mode: Use provided config, but validate required properties
		config = cnf.Config

		// Validate required properties
		if config.ID == "" {
			return nil, fmt.Errorf("required property 'ID' is missing")
		}
		if config.Secret == "" {
			return nil, fmt.Errorf("required property 'Secret' is missing")
		}
		if config.SecretKey == "" {
			return nil, fmt.Errorf("required property 'SecretKey' is missing")
		}
		if config.ServerURL == "" {
			return nil, fmt.Errorf("required property 'ServerURL' is missing")
		}
		if config.Branch == "" {
			return nil, fmt.Errorf("required property 'Branch' is missing")
		}

	default:
		return nil, fmt.Errorf("either URL or Config with required properties must be provided")
	}

	// Fetch and decrypt context data
	properties, err := clientService.GetContextDecryptedData(config)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch context data: %w", err)
	}
	if cnf.SetOSEnv {
		for _, contextData := range properties {
			os.Setenv(contextData.Property, contextData.Value.(string))
		}
	}

	return properties, nil
}
