package cli

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
)

func init() {
	cmd.Register(&Config)
}

// Config updates a config value with Ark.
var Config = cmd.Sub{
	Name:  "config",
	Alias: "c",
	Short: "Update an one of Ark's Configuration Values.",
	Args:  &ConfigArgs{},
	Run:   ConfigRun,
}

// ConfigArgs handles the specific arguments for the config command.
type ConfigArgs struct {
	Key   string
	Value []string `zero:"true"`
}

// ConfigRun updates one of Ark's internal config values.
func ConfigRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	args := c.Args.(*ConfigArgs)

	// Setup temporary configuration.
	cfg := config.Config{}

	// Parse configuration file.
	err := config.ParseFile(rFlags.Config, &cfg)
	checkError(rFlags, err)

	// Parse Category and Field Values
	confKeys := strings.Split(args.Key, ".")
	category := capName(confKeys[0])
	field := capName(confKeys[1])

	// Use reflect to get/set the value of the config struct
	reConf := reflect.ValueOf(&cfg).Elem().FieldByName(category)
	if len(args.Value) > 0 {
		reConf.FieldByName(field).SetString(args.Value[0])
		// Write changes back to config file.
		config.WriteFile(rFlags.Config, &cfg)
	} else {
		// Print current value if no new value given.
		fmt.Println(reConf.FieldByName(field))
	}
}

// capName capitalizes the first letter of a string
// and returns it.
func capName(in string) string {
	return strings.ToUpper(string(in[0])) + in[1:]
}
