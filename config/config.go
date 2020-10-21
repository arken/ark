package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/arkenproject/ait/utils"

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
	Remotes map[string]string
}

// ipfs defines the IPFS centric ait settings.
type ipfs struct {
	Path string
}

var (
	// Global is the configuration struct for the application.
	Global Config
	Path   string
)

// initialize the app config system. If a config doesn't exist, create one.
// If the config is out of date read the current config and rebuild with new fields.
func init() {
	// Determine the current user to build expected file path.
	user, err := user.Current()
	utils.CheckError(err)
	// Create expected config path.
	Path = filepath.Join(user.HomeDir, ".ait", "ait.config")
	readConf(&Global)
	// If the configuration version has changed update the config to the new
	// format while keeping the user's preferences.
	if Global.General.Version != defaultConf().General.Version {
		reloadConf()
		readConf(&Global)
	}
	ConsolidateEnvVars(&Global)

	err = createSwarmKey()
	if err != nil {
		log.Fatal(err)
	}
}

// Read the config or create a new one if it doesn't exist.
func readConf(conf *Config) {
	_, err := toml.DecodeFile(Path, &conf)
	if os.IsNotExist(err) {
		GenConf(defaultConf())
		genApplication(defaultApplication())
		readConf(conf)
	}
	if err != nil && !os.IsNotExist(err) {
		utils.FatalPrintln(err)
	}
}

func createSwarmKey() (err error) {
	keyData := []byte(`/key/swarm/psk/1.0.0/
/base16/
793bdb68b7cfd2f49071a299711df51f1c60283a047e4a8756a5c3a3d1ab776f`)

	os.MkdirAll(Global.IPFS.Path, os.ModePerm)
	err = ioutil.WriteFile(filepath.Join(Global.IPFS.Path, "swarm.key"), keyData, 0644)
	return err
}

// GetRemote takes a string and returns what should be a URL. If the string is
// a key in Global.Git.Remotes, then its value will be returned. If it itself
// a url, it will be returned untouched. This is to allow the arbitrary
// substitution of real remote URLs and remote aliases.
func GetRemote(remote string) string {
	url, ok := Global.Git.Remotes[remote]
	if ok {
		return url
	}
	return remote
}
