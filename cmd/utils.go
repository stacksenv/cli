package cmd

import (
	"errors"
	"log"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stacksenv/cli/pkg/homedir"
)

// Generate the replacements for all environment variables. This allows to
// use FB_BRANDING_DISABLE_EXTERNAL environment variables, even when the
// option name is branding.disableExternal.
func generateEnvKeyReplacements(cmd *cobra.Command) []string {
	replacements := []string{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		oldName := strings.ToUpper(f.Name)
		newName := strings.ToUpper(lo.SnakeCase(f.Name))
		replacements = append(replacements, oldName, newName)
	})

	return replacements
}

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()

	// Get config file from flag
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	// Configuration file
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

	// Environment variables
	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(generateEnvKeyReplacements(cmd)...))

	// Bind the flags
	err = v.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Read in configuration
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, viper.ConfigParseError{}) {
			return nil, err
		}

		log.Println("No config file used")
	} else {
		log.Printf("Using config file: %s", v.ConfigFileUsed())
	}

	// Return Viper
	return v, nil
}

type store struct {
	// *storage.Storage
	databaseExisted bool
}

type storeOptions struct {
	allowsNoDatabase bool
}

type cobraFunc func(cmd *cobra.Command, args []string) error

// withViperAndStore initializes Viper and the storage.Store and passes them to the callback function.
// This function should only be used by the root command. No other command should call
// this function directly.
func withViperAndStore(fn func(cmd *cobra.Command, args []string, v *viper.Viper, store *store) error, _ storeOptions) cobraFunc {
	return func(cmd *cobra.Command, args []string) error {
		v, err := initViper(cmd)
		if err != nil {
			return err
		}

		store := &store{
			// Storage:         storage,
			databaseExisted: false,
		}

		return fn(cmd, args, v, store)
	}
}
