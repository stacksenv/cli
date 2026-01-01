package stacksenv

// Config represents the configuration for connecting to a stacksenv server.
// It contains authentication credentials and server connection details.
type Config struct {
	ID           string `json:"id"`            // Unique identifier for the environment
	Secret       string `json:"secret"`        // Secret key for authentication
	SecretKey    string `json:"secretkey"`     // Additional secret key for encryption
	ServerURL    string `json:"serverurl"`     // Server hostname or IP address
	Branch       string `json:"branch"`        // Branch name (e.g., "dev", "prod")
	DisableHTTPS bool   `json:"disable_https"` // Whether to use HTTP instead of HTTPS
}

// ContextData represents a key-value pair for environment context data.
// It uses generics to support different value types.
type ContextData[T any] struct {
	Property string `json:"property"` // The property name (environment variable name)
	Value    T      `json:"value"`    // The property value
}

// ServerResponse represents the response structure from the stacksenv server.
type ServerResponse struct {
	Error         string `json:"error"` // Error message if request failed
	EncryptedData string `json:"data"`  // Encrypted data payload
}

// RequestConfig represents the configuration for a stacksenv request.
// It can contain either a URL to parse or a pre-configured Config struct.
type RequestConfig struct {
	URL      string  `json:"url"`    // Optional stacksenv URL to parse
	Config   *Config `json:"config"` // Optional pre-configured Config struct
	SetOSEnv bool    `json:"setenv"` // Whether to set OS environment variables
}
