package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stacksenv/cli/config"
	"github.com/stacksenv/cli/pkg/homedir"
	"go.yaml.in/yaml/v3"
)

// debugEnabled stores whether debug logging is enabled.
// It is set during viper initialization and used by all logging functions.
var debugEnabled bool

// debugLog prints a log message only if debug mode is enabled.
func debugLog(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf(format, v...)
	}
}

// debugLogLn prints a log message (without format) only if debug mode is enabled.
func debugLogLn(v ...interface{}) {
	if debugEnabled {
		log.Println(v...)
	}
}

// generateEnvKeyReplacements generates key replacement pairs for environment variable mapping.
// This allows environment variables like FB_BRANDING_DISABLE_EXTERNAL to map to configuration
// keys like branding.disableExternal by converting flag names to snake_case format.
func generateEnvKeyReplacements(cmd *cobra.Command) []string {
	replacements := []string{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		oldName := strings.ToUpper(f.Name)
		newName := strings.ToUpper(lo.SnakeCase(f.Name))
		replacements = append(replacements, oldName, newName)
	})

	return replacements
}

// loadConfigFile attempts to load a configuration file using viper and merge it into the main viper instance.
// It supports JSON and YAML formats, trying JSON first if the file has no extension.
// Returns true if the config was successfully loaded and merged.
func loadConfigFile(v *viper.Viper, configPath string, logMessage string) bool {
	vTemp := viper.New()
	vTemp.SetConfigFile(configPath)

	// If file has no extension, try JSON first, then YAML
	if filepath.Ext(configPath) == "" {
		vTemp.SetConfigType("json")
		if err := vTemp.ReadInConfig(); err != nil {
			vTemp.SetConfigType("yaml")
			if err := vTemp.ReadInConfig(); err != nil {
				return false
			}
		}
	} else {
		if err := vTemp.ReadInConfig(); err != nil {
			return false
		}
	}

	// Merge the loaded config into the main viper instance
	if err := v.MergeConfigMap(vTemp.AllSettings()); err != nil {
		return false
	}

	if logMessage != "" {
		debugLog(logMessage, configPath)
	}
	return true
}

// ensureGlobalConfigExists creates the global configuration file and directory if they don't exist.
// The config file is initialized with default values including serverurl from config.DefaultServerURL.
func ensureGlobalConfigExists(configPath string) error {
	configDir := filepath.Dir(configPath)

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	// Create .stacksenv directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create default config with serverurl and sessions properties
	defaultConfig := map[string]interface{}{
		"serverurl": config.DefaultServerURL,
		"sessions":  []interface{}{},
	}
	configJSON, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}
	configJSON = append(configJSON, '\n')

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return err
	}

	debugLog("Created global config file: %s", configPath)
	return nil
}

// getGlobalConfigPath returns the path to the global configuration file.
func getGlobalConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".stacksenv", "config"), nil
}

// readGlobalConfig reads the global configuration file and returns its contents.
// It supports both JSON and YAML formats and returns the data along with the detected format.
func readGlobalConfig() (map[string]interface{}, bool, error) {
	configPath, err := getGlobalConfigPath()
	if err != nil {
		return nil, false, err
	}

	configData := make(map[string]interface{})
	isYAML := false

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		// Read existing config
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, false, fmt.Errorf("failed to read config file: %w", err)
		}

		// Try to determine format and parse accordingly
		// Check if it's YAML (starts with key: or has YAML-like structure)
		if len(data) > 0 && (data[0] != '{' && data[0] != '[') {
			// Likely YAML format
			if err := yaml.Unmarshal(data, &configData); err == nil {
				isYAML = true
			} else {
				// Try JSON as fallback
				if err := json.Unmarshal(data, &configData); err != nil {
					return nil, false, fmt.Errorf("failed to parse config file (tried YAML and JSON): %w", err)
				}
			}
		} else {
			// Try JSON first
			if err := json.Unmarshal(data, &configData); err != nil {
				// Fallback to YAML
				if err := yaml.Unmarshal(data, &configData); err != nil {
					return nil, false, fmt.Errorf("failed to parse config file (tried JSON and YAML): %w", err)
				}
				isYAML = true
			}
		}
	} else {
		// Create default config structure if file doesn't exist
		configData = map[string]interface{}{
			"serverurl": config.DefaultServerURL,
			"sessions":  []interface{}{},
		}
		// Default to JSON format for new files
		isYAML = false
	}

	return configData, isYAML, nil
}

// writeGlobalConfig writes the configuration data to the global config file.
// It preserves the format (JSON or YAML) based on the isYAML parameter.
func writeGlobalConfig(configData map[string]interface{}, isYAML bool) error {
	configPath, err := getGlobalConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config back to file in the same format
	var configBytes []byte
	if isYAML {
		configBytes, err = yaml.Marshal(configData)
		if err != nil {
			return fmt.Errorf("failed to marshal config to YAML: %w", err)
		}
	} else {
		configBytes, err = json.MarshalIndent(configData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config to JSON: %w", err)
		}
		configBytes = append(configBytes, '\n')
	}

	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// updateGlobalConfig updates a property in the global configuration file.
// It reads the existing config, updates the specified key with the new value,
// and writes it back preserving the original format (JSON or YAML).
func updateGlobalConfig(key string, value interface{}) error {
	// Read existing config
	configData, isYAML, err := readGlobalConfig()
	if err != nil {
		return err
	}

	// Update the specified key
	configData[key] = value

	// Write updated config back
	if err := writeGlobalConfig(configData, isYAML); err != nil {
		return err
	}

	return nil
}

// createLocalConfig creates a local configuration file in the current working directory.
// The file is created as .stacksenv/config.json with default values.
// Returns an error if the file already exists or if creation fails.
func createLocalConfig() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	configDir := filepath.Join(cwd, ".stacksenv")
	configPath := filepath.Join(configDir, "config.json")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("local config file already exists: %s", configPath)
	}

	// Create .stacksenv directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default config with serverurl and sessions properties
	defaultConfig := map[string]interface{}{
		"serverurl": config.DefaultServerURL,
		"sessions":  []interface{}{},
	}
	configJSON, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configJSON = append(configJSON, '\n')

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// initViper initializes and configures a Viper instance with configuration from multiple sources.
// Configuration precedence (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (FB_ prefix)
// 3. Local project config (.stacksenv/config.json or .stacksenv/config.yaml in current directory)
// 4. Global user config ($HOME/.stacksenv/config)
// 5. System-wide config (/etc/stacksenv/.stacksenv)
// 6. Standard config paths (current directory, $HOME, /etc/stacksenv/)
func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()

	// Get config file path from command-line flag
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	// Configure config file search paths if no explicit config file is specified
	if cfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/stacksenv/")
		v.SetConfigName(".stacksenv")
	} else {
		v.SetConfigFile(cfgFile)
	}

	// Configure environment variable support
	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(generateEnvKeyReplacements(cmd)...))

	// Bind command-line flags to viper
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}

	// Get debug flag value and set global debugEnabled
	debugEnabled, _ = cmd.Flags().GetBool("debug")

	// Attempt to read configuration from standard paths
	configFound := false
	if err := v.ReadInConfig(); err != nil {
		var parseErr viper.ConfigParseError
		if errors.As(err, &parseErr) {
			return nil, err
		}
		debugLogLn("No config file used")
	} else {
		configFound = true
		debugLog("Using config file: %s", v.ConfigFileUsed())
	}

	// Load global fallback config if no config was found in standard paths
	if cfgFile == "" && !configFound {
		home, err := homedir.Dir()
		if err == nil {
			globalConfigPath := filepath.Join(home, ".stacksenv", "config")

			// Ensure global config file exists (create if missing)
			if err := ensureGlobalConfigExists(globalConfigPath); err != nil {
				debugLog("Failed to ensure global config exists: %v", err)
			}

			// Load and merge global config
			loadConfigFile(v, globalConfigPath, "Loaded config from: %s")
		}
	}

	// Load local project config (overwrites global config)
	// Priority: config.json > config.yaml > config.yml
	if cfgFile == "" {
		cwd, err := os.Getwd()
		if err == nil {
			stacksenvDir := filepath.Join(cwd, ".stacksenv")
			configFiles := []string{"config.json", "config.yaml", "config.yml"}

			for _, configFile := range configFiles {
				localConfigPath := filepath.Join(stacksenvDir, configFile)
				if _, err := os.Stat(localConfigPath); err == nil {
					if loadConfigFile(v, localConfigPath, "Loaded local config from: %s (overwrites global config)") {
						break // Only load the first found config file
					}
				}
			}
		}
	}

	return v, nil
}

// store represents the application's storage state.
// Currently contains only databaseExisted flag; storage field is reserved for future use.
type store struct {
	// Storage: *storage.Storage // Reserved for future implementation
	databaseExisted bool
}

// storeOptions contains options for configuring the store initialization.
type storeOptions struct {
	allowsNoDatabase bool
}

// cobraFunc is a type alias for Cobra command execution functions.
type cobraFunc func(cmd *cobra.Command, args []string) error

// withViperAndStore initializes Viper configuration and store, then passes them to the callback function.
// This wrapper ensures consistent initialization across commands and should only be used by the root command.
// Other commands should not call this function directly.
func withViperAndStore(fn func(cmd *cobra.Command, args []string, v *viper.Viper, store *store) error, _ storeOptions) cobraFunc {
	return func(cmd *cobra.Command, args []string) error {
		v, err := initViper(cmd)
		if err != nil {
			return err
		}

		store := &store{
			databaseExisted: false,
		}

		return fn(cmd, args, v, store)
	}
}

// marshal writes data to a file in the format determined by the file extension.
// Supported formats: JSON (with indentation), YAML, YML.
// Returns an error if the format is unsupported or file operations fail.
func marshal(filename string, data interface{}) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		return encoder.Encode(data)
	case ".yml", ".yaml":
		encoder := yaml.NewEncoder(fd)
		return encoder.Encode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

// unmarshal reads and decodes data from a file based on its extension.
// Supported formats: JSON, YAML, YML.
// Returns an error if the format is unsupported or file operations fail.
func unmarshal(filename string, data interface{}) error {
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		return json.NewDecoder(fd).Decode(data)
	case ".yml", ".yaml":
		return yaml.NewDecoder(fd).Decode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

// jsonYamlArg validates that exactly one argument is provided and that it has
// a supported file extension (.json, .yml, or .yaml).
// Returns an error if validation fails.
func jsonYamlArg(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	switch ext := filepath.Ext(args[0]); ext {
	case ".json", ".yml", ".yaml":
		return nil
	default:
		return errors.New("invalid format: " + ext)
	}
}

// convertCmdStrToCmdArray converts a command string to an array of command arguments.
// Trims whitespace and splits by spaces. Returns an empty array if the input is blank
// (whitespace-only), ensuring the result is never []string{""}.
func convertCmdStrToCmdArray(cmd string) []string {
	var cmdArray []string
	trimmedCmdStr := strings.TrimSpace(cmd)
	if trimmedCmdStr != "" {
		cmdArray = strings.Split(trimmedCmdStr, " ")
	}
	return cmdArray
}
