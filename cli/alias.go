package cli

import (
	"fmt"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
)

func init() {
	cmd.Register(&Alias)
}

// Alias creates a custom alias mapping in Ark.
var Alias = cmd.Sub{
	Name:  "alias",
	Alias: "a",
	Short: "Create a shortcut for a manifest URL.",
	Args:  &AliasArgs{},
	Flags: &AliasFlags{},
	Run:   AliasRun,
}

// AliasArgs handles the specific arguments for the alias command.
type AliasArgs struct {
	Shortcut string
	URL      []string `zero:"true"`
}

// AliasFlags handles the specific flags for the alias command.
type AliasFlags struct {
	Delete bool `short:"d" long:"delete" desc:"delete an alias shortcut."`
}

// AliasRun creates a custom alias mapping for a manifest url.
func AliasRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	args := c.Args.(*AliasArgs)

	flags := c.Flags.(*AliasFlags)

	// Setup temporary configuration.
	cfg := config.Config{}

	// Parse configuration file.
	err := config.ParseFile(rFlags.Config, &cfg)
	checkError(rFlags, err)

	if len(args.URL) > 0 {
		cfg.Manifest.Aliases[args.Shortcut] = args.URL[0]
		// Write changes back to config file.
		config.WriteFile(rFlags.Config, &cfg)
	} else {
		if flags.Delete {
			delete(cfg.Manifest.Aliases, args.Shortcut)
			// Write changes back to config file.
			config.WriteFile(rFlags.Config, &cfg)
		} else {
			fmt.Println(cfg.Manifest.Aliases[args.Shortcut])
		}
	}
}
