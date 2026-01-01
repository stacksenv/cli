package stacksenv

import (
	"fmt"
	"strings"
)

// DefaultURLParser is the default implementation of URLParser.
type DefaultURLParser struct{}

// NewURLParser creates a new instance of DefaultURLParser.
func NewURLParser() URLParser {
	return &DefaultURLParser{}
}

// ParseURL parses a stacksenv URL string and returns a Config.
//
// URL format: stacksenv://ID:SECRET:SECRET_KEY@SERVER_URL/BRANCH?disable_https=true
//
// Example: stacksenv://abc123:secret:key@example.com/dev?disable_https=false
//
// Returns an error if the URL format is invalid.
func (p *DefaultURLParser) ParseURL(urlStr string) (Config, error) {
	config := Config{}

	// Split URL into credentials and server parts
	parts := strings.Split(urlStr, "@")
	if len(parts) != 2 {
		return config, fmt.Errorf("invalid URL format: expected 'credentials@server', got: %s", urlStr)
	}

	// Parse credentials: ID:SECRET:SECRET_KEY
	credParts := strings.Split(parts[0], ":")
	if len(credParts) != 3 {
		return config, fmt.Errorf("invalid credentials format: expected 'ID:SECRET:SECRET_KEY', got: %s", parts[0])
	}
	config.ID = credParts[0]
	config.Secret = credParts[1]
	config.SecretKey = credParts[2]

	// Parse server and branch: SERVER_URL/BRANCH
	serverParts := strings.Split(parts[1], "/")
	if len(serverParts) != 2 {
		return config, fmt.Errorf("invalid server URL format: expected 'SERVER_URL/BRANCH', got: %s", parts[1])
	}
	config.ServerURL = serverParts[0]

	// Parse branch and query parameters: BRANCH?disable_https=true
	branchAndOptions := strings.Split(serverParts[1], "?")
	if len(branchAndOptions) == 0 {
		return config, fmt.Errorf("invalid branch format: %s", serverParts[1])
	}
	config.Branch = branchAndOptions[0]

	// Parse query parameters
	if len(branchAndOptions) > 1 {
		options := strings.Split(branchAndOptions[1], "&")
		for _, option := range options {
			optionParts := strings.Split(option, "=")
			if len(optionParts) != 2 {
				return config, fmt.Errorf("invalid query parameter format: %s", option)
			}
			if optionParts[0] == "disable_https" {
				config.DisableHTTPS = optionParts[1] == "true"
			}
		}
	}

	return config, nil
}

// ParseURL is a convenience function that uses the default parser.
// It's maintained for backward compatibility.
func ParseURL(urlStr string) (Config, error) {
	parser := NewURLParser()
	return parser.ParseURL(urlStr)
}
