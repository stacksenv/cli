package stacksenv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ServerResponse struct {
	Error         string `json:"error"`
	EncryptedData string `json:"data"`
}

func SendCLIRequest(config *Config) (*http.Response, error) {
	baseURL := ""
	if config.DisableHTTPS {
		baseURL = fmt.Sprintf("http://%s/cli", config.ServerURL)
	} else {
		baseURL = fmt.Sprintf("https://%s/cli", config.ServerURL)
	}

	// Build URL with query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := url.Values{}
	params.Add("id", config.ID)
	params.Add("branch", config.Branch)

	u.RawQuery = params.Encode()

	// Send GET request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}

	return resp, nil
}

func GetContextDecryptedData(cnf *Config) ([]ContextData[any], error) {
	var result []ContextData[any]
	resp, err := SendCLIRequest(cnf)
	if err != nil {
		return result, err
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
	// Parse and format as JSON
	var jsonData map[string]any
	if err := json.Unmarshal(body, &jsonData); err != nil {
		// If not valid JSON, print raw response
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	errMsg, ok := jsonData["error"].(string)
	if ok {
		return result, fmt.Errorf("error: %s", errMsg)
	}

	encryptedData, ok := jsonData["data"].(string)
	if ok {
		return Decrypt(encryptedData, cnf.Secret, cnf.SecretKey)
	}

	return result, fmt.Errorf("no encrypted data found")
}
