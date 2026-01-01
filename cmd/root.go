package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// var (
// 	flagNamesMigrations = map[string]string{}

// 	warnedFlags = map[string]bool{}
// )

// TODO(remove): remove after July 2026.
// func migrateFlagNames(_ *pflag.FlagSet, name string) pflag.NormalizedName {
// 	if newName, ok := flagNamesMigrations[name]; ok {

// 		if !warnedFlags[name] {
// 			warnedFlags[name] = true
// 			log.Printf("DEPRECATION NOTICE: Flag --%s has been deprecated, use --%s instead\n", name, newName)
// 		}

// 		name = newName
// 	}

// 	return pflag.NormalizedName(name)
// }

func init() {
	rootCmd.SilenceUsage = true
	// rootCmd.SetGlobalNormalizationFunc(migrateFlagNames)

	rootCmd.SetVersionTemplate("Stacksenv version {{printf \"%s\" .Version}}\n")

	// Flags available across the whole program
	persistent := rootCmd.PersistentFlags()
	persistent.StringP("config", "c", "", "config file path")
	persistent.BoolP("debug", "d", false, "enable debug logging")
}

var rootCmd = &cobra.Command{
	Use:   "stacksenv",
	Short: "Stacksenv is a CLI for managing your Environment Variables",
	Long: `Stacksenv is a CLI for managing your Environment Variables

If "--config" is not specified, Stacksenv will look for a configuration
file named .stacksenv.{json, toml, yaml, yml} in the following directories:

- ./
- $HOME/
- /etc/stacksenv/

**Note:** Only the options listed below can be set via the config file or
environment variables. Other configuration options live exclusively in the
environment variables and so they must be set by the "env set" or "env
import" commands.

The precedence of the configuration values are as follows:

- Flags
- Environment variables
- Configuration file
- Environment variables
- Defaults

Also, if the environment variables path doesn't exist, Stacksenv will enter into
the quick setup mode and a new environment variables will be bootstrapped and a new
user created with the credentials from options "username" and "password".`,
	RunE: withViperAndStore(func(_ *cobra.Command, _ []string, v *viper.Viper, _ *store) error {
		fmt.Println("Hello, World!: ", v.GetInt("port"))
		fmt.Println("Hello, World!: ", v.GetString("serverurl"))

		return nil
	}, storeOptions{allowsNoDatabase: true}),
}
