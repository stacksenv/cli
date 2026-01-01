package stacksenv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DefaultHTTPClient is the default implementation of HTTPClient using net/http.
type DefaultHTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with default settings.
// For better performance, it reuses connections and sets reasonable timeouts.
func NewHTTPClient() HTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

// Do sends an HTTP request and returns an HTTP response.
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// DefaultClientService is the default implementation of ClientService.
type DefaultClientService struct {
	httpClient HTTPClient
	crypto     CryptoService
}

// NewClientService creates a new client service with the provided dependencies.
func NewClientService(httpClient HTTPClient, crypto CryptoService) ClientService {
	return &DefaultClientService{
		httpClient: httpClient,
		crypto:     crypto,
	}
}

// SendCLIRequest sends a GET request to the stacksenv server to fetch context data.
//
// It constructs the URL with the appropriate protocol (HTTP/HTTPS) based on config.DisableHTTPS,
// and includes the ID and branch as query parameters.
//
// Returns the HTTP response or an error if the request fails.
func SendCLIRequest(config *Config, httpClient HTTPClient) (*http.Response, error) {
	// Determine protocol
	protocol := "https"
	if config.DisableHTTPS {
		protocol = "http"
	}

	// Build base URL
	baseURL := fmt.Sprintf("%s://%s/cli", protocol, config.ServerURL)

	// Parse and build URL with query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := url.Values{}
	params.Set("id", config.ID)
	params.Set("branch", config.Branch)
	u.RawQuery = params.Encode()

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}

	return resp, nil
}

// GetContextDecryptedData fetches encrypted context data from the server and decrypts it.
//
// The process:
//  1. Sends a GET request to the server with ID and branch parameters
//  2. Reads and parses the JSON response
//  3. Extracts the encrypted data payload
//  4. Decrypts the data using the provided secret and secret key
//  5. Returns the decrypted context data as a slice of ContextData
//
// Returns an error if any step fails (HTTP request, JSON parsing, or decryption).
func (s *DefaultClientService) GetContextDecryptedData(config *Config) ([]ContextData[any], error) {
	var result []ContextData[any]

	// Send request to server
	resp, err := SendCLIRequest(config, s.httpClient)
	if err != nil {
		return result, fmt.Errorf("failed to send CLI request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var jsonData map[string]any
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Check for error in response
	if errMsg, ok := jsonData["error"].(string); ok && errMsg != "" {
		return result, fmt.Errorf("server error: %s", errMsg)
	}

	// Extract encrypted data
	encryptedData, ok := jsonData["data"].(string)
	if !ok || encryptedData == "" {
		return result, fmt.Errorf("no encrypted data found in response")
	}

	// Decrypt data - try multiple combinations to match server encryption
	// The server encryption format may vary, so we try common patterns in order of likelihood

	// Try 1: SecretKey as shared secret, Secret|SecretKey as AAD (most common pattern)
	aad := fmt.Sprintf("%s|%s", config.Secret, config.SecretKey)
	if result, err := s.crypto.Decrypt(encryptedData, config.SecretKey, aad); err == nil {
		return result, nil
	}

	// Try 2: Secret as shared secret, SecretKey as AAD
	if result, err := s.crypto.Decrypt(encryptedData, config.Secret, config.SecretKey); err == nil {
		return result, nil
	}

	// Try 3: SecretKey as shared secret, Secret as AAD
	if result, err := s.crypto.Decrypt(encryptedData, config.SecretKey, config.Secret); err == nil {
		return result, nil
	}

	// Try 4: Secret as shared secret, Secret|SecretKey as AAD
	if result, err := s.crypto.Decrypt(encryptedData, config.Secret, aad); err == nil {
		return result, nil
	}

	// Try 5: SecretKey as shared secret, empty AAD
	if result, err := s.crypto.Decrypt(encryptedData, config.SecretKey, ""); err == nil {
		return result, nil
	}

	// Try 6: Secret as shared secret, empty AAD
	if result, err := s.crypto.Decrypt(encryptedData, config.Secret, ""); err == nil {
		return result, nil
	}

	// If all attempts fail, return error from the most likely combination
	return nil, fmt.Errorf("failed to decrypt data: tried all common encryption key/AAD combinations")
}

// GetContextDecryptedData is a convenience function that uses default implementations.
// It's maintained for backward compatibility.
func GetContextDecryptedData(config *Config) ([]ContextData[any], error) {
	httpClient := NewHTTPClient()
	crypto := NewCryptoService()
	service := NewClientService(httpClient, crypto)
	return service.GetContextDecryptedData(config)
}
