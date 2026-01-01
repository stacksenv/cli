# Stacksenv Package

The `stacksenv` package provides a professional, well-structured implementation for handling stacksenv URLs, fetching encrypted context data from servers, and executing commands with environment variables.

## Overview

This package implements a clean architecture with interfaces for dependency injection, making it highly testable and maintainable. It handles:

- URL parsing for stacksenv protocol
- HTTP communication with stacksenv servers
- Encryption/decryption of context data
- Command execution with environment variables

## Quick Start

### Using as a Go Library

The package provides two main functions for different use cases:

1. **`HandleStacksENV`** - For library usage: Fetches and returns context data (properties)
2. **`HandleStacksenvURLCLI`** - For CLI usage: Fetches context data and executes commands

### Library Usage - Fetching Context Data

The simplest way to use this package as a library is through `HandleStacksENV`:

```go
package main

import (
    "fmt"
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    // Option 1: Using URL
    config := &stacksenv.RequestConfig{
        URL: "stacksenv://ID:SECRET:SECRET_KEY@server.com/dev",
        SetOSEnv: true, // Optionally set environment variables in OS
    }
    
    properties, err := stacksenv.HandleStacksENV(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the properties
    for _, prop := range properties {
        fmt.Printf("%s = %v\n", prop.Property, prop.Value)
    }
    
    // Option 2: Using Config struct
    config2 := &stacksenv.RequestConfig{
        Config: &stacksenv.Config{
            ID:        "abc123",
            Secret:    "mysecret",
            SecretKey: "mykey",
            ServerURL: "api.example.com",
            Branch:    "dev",
        },
        SetOSEnv: false, // Don't set OS env, just return properties
    }
    
    properties2, err := stacksenv.HandleStacksENV(config2)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access environment variables from properties
    for _, prop := range properties2 {
        fmt.Printf("Environment variable: %s\n", prop.Property)
    }
}
```

### CLI Usage - Execute Commands

For CLI applications that need to execute commands with environment variables:

```go
package main

import (
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    // Handle a stacksenv URL and execute a command
    url := "stacksenv://ID:SECRET:SECRET_KEY@server.com/dev"
    args := []string{"node", "-v"}
    
    if err := stacksenv.HandleStacksenvURLCLI(url, args); err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Usage with Dependency Injection

For more control and testability, use the `Handler` struct directly:

```go
package main

import (
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    // Create a handler with default implementations
    handler := stacksenv.NewHandler(nil, nil, nil)
    
    // Or create with custom implementations
    httpClient := stacksenv.NewHTTPClient()
    crypto := stacksenv.NewCryptoService()
    clientService := stacksenv.NewClientService(httpClient, crypto)
    
    handler := stacksenv.NewHandler(nil, clientService, nil)
    
    // Use the handler
    url := "stacksenv://ID:SECRET:SECRET_KEY@server.com/dev"
    args := []string{"python", "script.py"}
    
    if err := handler.HandleStacksenvURLCLI(url, args); err != nil {
        log.Fatal(err)
    }
}
```

## URL Format

The stacksenv URL follows this format:

```
stacksenv://ID:SECRET:SECRET_KEY@SERVER_URL/BRANCH?disable_https=true
```

### Components

- **ID**: Unique identifier for the environment
- **SECRET**: Secret key for authentication
- **SECRET_KEY**: Additional secret key for encryption/decryption
- **SERVER_URL**: Server hostname or IP address (e.g., `example.com` or `10.0.0.1`)
- **BRANCH**: Branch name (e.g., `dev`, `prod`, `staging`)
- **disable_https**: Optional query parameter (`true`/`false`) to use HTTP instead of HTTPS

### Examples

```go
// Basic URL with HTTPS
url1 := "stacksenv://abc123:secret:key@api.example.com/prod"

// URL with HTTP (disable HTTPS)
url2 := "stacksenv://abc123:secret:key@10.0.0.1/dev?disable_https=true"

// URL for development branch
url3 := "stacksenv://xyz789:mysecret:mykey@localhost:8080/dev"
```

## Function Reference

### HandleStacksENV

The main function for library usage. Fetches and returns context data (properties) from the stacksenv server.

```go
func HandleStacksENV(cnf *RequestConfig) ([]ContextData[any], error)
```

**Parameters:**
- `cnf`: Configuration struct containing either a URL or Config with required properties

**Returns:**
- `[]ContextData[any]`: Slice of context data (properties) fetched from the server
- `error`: An error if URL parsing, validation, or data fetching fails

**RequestConfig Fields:**
- `URL` (string): Optional stacksenv URL. If provided, it will be parsed to extract configuration
- `Config` (*Config): Optional Config struct. Required if URL is not provided
- `SetOSEnv` (bool): If true, sets environment variables in the OS environment using `os.Setenv()`

**Modes of Operation:**

1. **URL Mode**: If `URL` is provided, it parses the URL and uses the extracted configuration
2. **Config Mode**: If `URL` is empty, it validates and uses the provided `Config` struct
3. **OS Environment**: If `SetOSEnv` is true, environment variables are set in the OS environment

**Required Properties for Config Mode:**
- `ID`: Unique identifier for the environment
- `Secret`: Secret key for authentication
- `SecretKey`: Additional secret key for encryption
- `ServerURL`: Server hostname or IP address
- `Branch`: Branch name (e.g., "dev", "prod")

**Example:**

```go
// Using URL with OS environment variables
config := &stacksenv.RequestConfig{
    URL:      "stacksenv://abc123:secret:key@api.example.com/prod",
    SetOSEnv: true, // Sets os.Setenv() for each property
}

properties, err := stacksenv.HandleStacksENV(config)
if err != nil {
    return err
}

// Properties are now available, and if SetOSEnv=true, they're also in os.Getenv()
for _, prop := range properties {
    fmt.Printf("%s = %v\n", prop.Property, prop.Value)
    // If SetOSEnv=true, you can also access via: os.Getenv(prop.Property)
}

// Using Config struct without setting OS environment
config2 := &stacksenv.RequestConfig{
    Config: &stacksenv.Config{
        ID:        "xyz789",
        Secret:    "mysecret",
        SecretKey: "mykey",
        ServerURL: "10.0.0.1",
        Branch:    "dev",
        DisableHTTPS: true,
    },
    SetOSEnv: false, // Just return properties, don't set OS env
}

properties2, err := stacksenv.HandleStacksENV(config2)
if err != nil {
    return err
}

// Use properties directly without OS environment variables
dbURL := ""
for _, prop := range properties2 {
    if prop.Property == "db_url" {
        dbURL = prop.Value.(string)
        break
    }
}
```

### HandleStacksenvURLCLI

The convenience function for CLI usage. Handles stacksenv URLs and executes commands with environment variables.

```go
func HandleStacksenvURLCLI(url string, args []string) error
```

**Parameters:**
- `url`: The stacksenv URL (format: `stacksenv://ID:SECRET:SECRET_KEY@SERVER_URL/BRANCH`)
- `args`: Command and arguments to execute (e.g., `["node", "-v"]` or `["python", "script.py"]`)

**Returns:**
- `error`: An error if URL parsing, data fetching, or command execution fails

**Example:**

```go
// Execute a simple command
err := stacksenv.HandleStacksenvURLCLI(
    "stacksenv://abc123:secret:key@server.com/dev",
    []string{"echo", "$DB_URL"},
)

// Execute with shell command
err := stacksenv.HandleStacksenvURLCLI(
    "stacksenv://abc123:secret:key@server.com/dev",
    []string{"sh", "-c", "echo $DB_URL"},
)

// Execute without URL (just run command)
err := stacksenv.HandleStacksenvURLCLI("", []string{"ls", "-la"})
```

### NewHandler

Creates a new Handler with the provided dependencies. If `nil` is passed for any dependency, a default implementation will be used.

```go
func NewHandler(
    urlParser URLParser,
    clientService ClientService,
    commandExecutor CommandExecutor,
) *Handler
```

**Example:**

```go
// Use all defaults
handler := stacksenv.NewHandler(nil, nil, nil)

// Custom URL parser
customParser := &MyCustomParser{}
handler := stacksenv.NewHandler(customParser, nil, nil)

// Custom client service with custom HTTP client
httpClient := &MyHTTPClient{}
crypto := stacksenv.NewCryptoService()
clientService := stacksenv.NewClientService(httpClient, crypto)
handler := stacksenv.NewHandler(nil, clientService, nil)
```

## Architecture

### Interfaces

The package defines several interfaces for dependency injection:

- **`HTTPClient`**: Interface for making HTTP requests
- **`URLParser`**: Interface for parsing stacksenv URLs
- **`CryptoService`**: Interface for encryption/decryption operations
- **`CommandExecutor`**: Interface for executing system commands
- **`ClientService`**: Interface for fetching context data from the server

### Default Implementations

- **`DefaultHTTPClient`**: Uses `net/http` with connection pooling
- **`DefaultURLParser`**: Parses stacksenv URL format
- **`DefaultCryptoService`**: AES-256-GCM encryption/decryption
- **`DefaultCommandExecutor`**: Executes commands using `os/exec`
- **`DefaultClientService`**: Fetches and decrypts context data

## How It Works

### HandleStacksENV (Library Usage)

1. **URL Parsing** (if URL provided): The URL is parsed to extract configuration (ID, Secret, SecretKey, ServerURL, Branch)
2. **Config Validation** (if Config provided): Validates that all required properties are present
3. **Context Data Fetching**: 
   - Sends a GET request to `{protocol}://{ServerURL}/cli?id={ID}&branch={Branch}`
   - Receives encrypted JSON response
   - Decrypts the data using SecretKey as the encryption key and Secret as AAD
4. **OS Environment Setup** (if SetOSEnv=true): 
   - Calls `os.Setenv()` for each property
   - Makes environment variables available to the current process
5. **Return Properties**: Returns the context data as a slice of `ContextData[any]`

### HandleStacksenvURLCLI (CLI Usage)

1. **URL Parsing**: The URL is parsed to extract configuration (ID, Secret, SecretKey, ServerURL, Branch)
2. **Context Data Fetching**: Same as above
3. **Environment Variable Setup**: 
   - Converts decrypted context data to environment variables
   - Each property becomes an environment variable (e.g., `db_url=postgresql://...`)
4. **Command Execution**: 
   - Executes the provided command with the environment variables
   - Command has access to all the context data as environment variables

## Examples

### Example 1: Library Usage - Fetch Properties

```go
package main

import (
    "fmt"
    "log"
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    // Fetch properties using URL
    config := &stacksenv.RequestConfig{
        URL: "stacksenv://myid:mysecret:mykey@api.example.com/prod",
    }
    
    properties, err := stacksenv.HandleStacksENV(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use properties in your application
    for _, prop := range properties {
        fmt.Printf("%s = %v\n", prop.Property, prop.Value)
    }
}
```

### Example 2: Library Usage - Set OS Environment Variables

```go
package main

import (
    "os"
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    // Fetch and set OS environment variables
    config := &stacksenv.RequestConfig{
        URL:      "stacksenv://myid:mysecret:mykey@api.example.com/dev",
        SetOSEnv: true, // This will call os.Setenv() for each property
    }
    
    properties, err := stacksenv.HandleStacksENV(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Now you can access via os.Getenv()
    dbURL := os.Getenv("db_url")
    apiKey := os.Getenv("api_key")
    
    // Or use the properties directly
    for _, prop := range properties {
        // Both are available if SetOSEnv=true
        fmt.Printf("Property: %s\n", prop.Property)
    }
}
```

### Example 3: Library Usage - Using Config Struct

```go
package main

import (
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    config := &stacksenv.RequestConfig{
        Config: &stacksenv.Config{
            ID:           "abc123",
            Secret:       "mysecret",
            SecretKey:    "mykey",
            ServerURL:    "api.example.com",
            Branch:       "prod",
            DisableHTTPS: false,
        },
        SetOSEnv: true,
    }
    
    properties, err := stacksenv.HandleStacksENV(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access specific property
    for _, prop := range properties {
        if prop.Property == "database_url" {
            dbURL := prop.Value.(string)
            // Use dbURL in your application
        }
    }
}
```

### Example 4: CLI Usage - Execute Commands

```go
package main

import (
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    url := "stacksenv://myid:mysecret:mykey@api.example.com/prod"
    args := []string{"node", "-v"}
    
    if err := stacksenv.HandleStacksenvURLCLI(url, args); err != nil {
        log.Fatal(err)
    }
}
```

### Example 5: CLI Usage - Using Environment Variables in Scripts

```go
package main

import (
    "github.com/stacksenv/cli/pkg/stacksenv"
)

func main() {
    url := "stacksenv://myid:mysecret:mykey@api.example.com/dev"
    // The script can access $DB_URL, $API_KEY, etc. from the context data
    args := []string{"bash", "-c", "echo $DB_URL && python app.py"}
    
    stacksenv.HandleStacksenvURLCLI(url, args)
}
```

### Example 6: Custom HTTP Client for Testing

```go
package main

import (
    "net/http"
    "github.com/stacksenv/cli/pkg/stacksenv"
)

type MockHTTPClient struct{}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    // Return mock response for testing
}

func main() {
    mockClient := &MockHTTPClient{}
    crypto := stacksenv.NewCryptoService()
    clientService := stacksenv.NewClientService(mockClient, crypto)
    
    handler := stacksenv.NewHandler(nil, clientService, nil)
    handler.HandleStacksenvURLCLI("stacksenv://...", []string{"echo", "test"})
}
```

## Error Handling

The function returns errors in the following cases:

- **URL Parsing Errors**: Invalid URL format
- **HTTP Errors**: Network issues, server errors, or invalid responses
- **Decryption Errors**: Failed to decrypt data (wrong keys or corrupted data)
- **Command Execution Errors**: Command failed to execute

Example error handling:

```go
err := stacksenv.HandleStacksenvURLCLI(url, args)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed to parse"):
        fmt.Println("Invalid URL format")
    case strings.Contains(err.Error(), "HTTP error"):
        fmt.Println("Server communication error")
    case strings.Contains(err.Error(), "decrypt/auth failed"):
        fmt.Println("Authentication failed - check your credentials")
    case strings.Contains(err.Error(), "failed to execute command"):
        fmt.Println("Command execution failed")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## Security Considerations

1. **Credentials**: Never log or expose Secret and SecretKey values
2. **HTTPS**: Always use HTTPS in production (set `disable_https=false` or omit the parameter)
3. **Environment Variables**: Sensitive data is passed as environment variables to child processes
4. **Encryption**: Data is encrypted using AES-256-GCM with authenticated encryption

## Package Structure

```
pkg/stacksenv/
├── types.go          # Type definitions (Config, ContextData, ServerResponse)
├── interfaces.go     # Interface definitions for dependency injection
├── utils.go          # URL parsing utilities
├── http.go           # HTTP client and client service
├── crypt.go          # Encryption/decryption service
└── stackenv.go       # Main handler and command executor
```

## Testing

The package is designed with interfaces to facilitate testing. You can create mock implementations of any interface for unit testing:

```go
type MockClientService struct{}

func (m *MockClientService) GetContextDecryptedData(config *stacksenv.Config) ([]stacksenv.ContextData[any], error) {
    return []stacksenv.ContextData[any]{
        {Property: "db_url", Value: "postgresql://localhost/db"},
    }, nil
}

// Use in tests
handler := stacksenv.NewHandler(nil, &MockClientService{}, nil)
```

## License

This package is part of the stacksenv CLI project.

