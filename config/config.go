package config

import (
	"github.com/arkenproject/ait/utils"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config defines the configuration struct for importing settings from TOML.
type Config struct {
	General general
	Git     git
	IPFS    ipfs
}

// general defines the substruct about general application settings.
type general struct {
	Version string
	Editor  string
}

// git defines git specific config settings.
type git struct {
	Name  string
	Email string
}

// ipfs defines the IPFS centric ait settings.
type ipfs struct {
	Path string
}

var (
	// Global is the configuration struct for the application.
	Global Config
	path   string
)

// initialize the app config system. If a config doesn't exist, create one.
// If the config is out of date read the current config and rebuild with new fields.
func init() {
	// Determine the current user to build expected file path.
	user, err := user.Current()
	utils.CheckError(err)
	// Create expected config path.
	path = filepath.Join(user.HomeDir, ".ait", "ait.config")
	readConf(&Global)
	// If the configuration version has changed update the config to the new
	// format while keeping the user's preferences.
	if Global.General.Version != defaultConf().General.Version {
		reloadConf()
		readConf(&Global)
	}
	ConsolidateEnvVars(&Global)
}

// Read the config or create a new one if it doesn't exist.
func readConf(conf *Config) {
	_, err := toml.DecodeFile(path, &conf)
	if os.IsNotExist(err) {
		genConf(defaultConf())
		readConf(conf)
	}
	if err != nil && !os.IsNotExist(err) {
		utils.FatalPrintln(err)
	}
}
