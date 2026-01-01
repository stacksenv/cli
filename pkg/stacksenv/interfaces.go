package stacksenv

import (
	"net/http"
)

// HTTPClient defines the interface for making HTTP requests.
// This abstraction allows for easier testing and custom HTTP client implementations.
type HTTPClient interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
}

// URLParser defines the interface for parsing stacksenv URLs.
type URLParser interface {
	// ParseURL parses a stacksenv URL string and returns a Config.
	// The URL format is: stacksenv://ID:SECRET:SECRET_KEY@SERVER_URL/BRANCH?disable_https=true
	ParseURL(url string) (Config, error)
}

// CryptoService defines the interface for encryption and decryption operations.
type CryptoService interface {
	// Encrypt encrypts a slice of context data using the provided secret and AAD.
	Encrypt(data []ContextData[any], sharedSecret, aad string) (string, error)

	// Decrypt decrypts an encrypted string and returns the context data.
	Decrypt(encrypted string, sharedSecret, aad string) ([]ContextData[any], error)
}

// CommandExecutor defines the interface for executing system commands.
type CommandExecutor interface {
	// Execute runs a command with the given arguments and environment variables.
	// It returns an error if the command execution fails.
	Execute(command string, args []string, env []string) error
}

// ClientService defines the interface for fetching context data from the server.
type ClientService interface {
	// GetContextDecryptedData fetches and decrypts context data from the server.
	GetContextDecryptedData(config *Config) ([]ContextData[any], error)
}
