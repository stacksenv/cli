package stackenv

import (
	"fmt"
	"strings"
)

type Config struct {
	ID           string `json:"id"`
	Secret       string `json:"secret"`
	SecretKey    string `json:"secretkey"`
	ServerURL    string `json:"serverurl"`
	Branch       string `json:"branch"`
	DisableHTTPS bool   `json:"disable_https"`
}

func ParseURL(url string) (Config, error) {
	config := Config{}

	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return config, fmt.Errorf("invalid URL: %s", url)
	}
	idAndKeyAndSecretKey := strings.Split(parts[0], ":")
	if len(idAndKeyAndSecretKey) != 3 {
		return config, fmt.Errorf("invalid ID and Key or Secret Key: %s", parts[0])
	}
	config.ID = idAndKeyAndSecretKey[0]

	config.Secret = idAndKeyAndSecretKey[1]
	config.SecretKey = idAndKeyAndSecretKey[2]

	parts = strings.Split(parts[1], "/")
	if len(parts) != 2 {
		return config, fmt.Errorf("invalid Server URL: %s", parts[1])
	}
	config.ServerURL = parts[0]

	brenchAndOptions := strings.Split(parts[1], "?")
	if len(brenchAndOptions) == 0 {
		return config, fmt.Errorf("invalid Server URL: %s", parts[2])
	}
	config.Branch = brenchAndOptions[0]

	if len(brenchAndOptions) > 1 {
		options := strings.Split(brenchAndOptions[1], "&")
		for _, option := range options {
			optionParts := strings.Split(option, "=")
			if len(optionParts) != 2 {
				return config, fmt.Errorf("invalid Server URL: %s", option)
			}
			if optionParts[0] == "disable_https" {
				config.DisableHTTPS = true
			}
		}

	}

	return config, nil
}
