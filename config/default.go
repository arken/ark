package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// defaultConf defines the default values for AIT's configuration.
func defaultConf() Config {
	result := Config{
		General: general{
			// Configuration version number. If a field is added or changed
			// in this default, the version must be changed to tell the app
			// to rebuild the users config files.
			Version: "0.0.1",
		},
		Git: git{
			Name:  "",
			Email: "",
		},
	}
	return result
}

// genConf encodes the values of the Config stuct back into a TOML file.
func genConf(conf Config) {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(conf)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
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
