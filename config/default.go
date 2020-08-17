package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/utils"

	"github.com/BurntSushi/toml"
)

// defaultConf defines the default values for AIT's configuration.
func defaultConf() Config {
	result := Config{
		General: general{
			// Configuration version number. If a field is added or changed
			// in this default, the version must be changed to tell the app
			// to rebuild the users config files.
			Version: "0.0.3",
			Editor:  "nano",
		},
		Git: git{
			Name:  "",
			Email: "",
		},
		IPFS: ipfs{
			Path: filepath.Join(filepath.Dir(path), "ipfs"),
		},
	}
	return result
}

// genConf encodes the values of the Config stuct back into a TOML file.
func genConf(conf Config) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(conf)
	utils.CheckError(err)
	f, err := os.Create(path)
	utils.CheckError(err)
	defer f.Close()
	f.Write(buf.Bytes())
}

// reloadConf imports the users config onto a default config and then rewrites
// the configuration file.
func reloadConf() {
	result := defaultConf()
	readConf(&result)
	result.General.Version = defaultConf().General.Version
	genConf(result)
}
