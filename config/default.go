package config

import (
	"bytes"
	"io/ioutil"
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
			Path: filepath.Join(filepath.Dir(Path), "ipfs"),
		},
	}
	return result
}

// genConf encodes the values of the Config struct back into a TOML file.
func genConf(conf Config, appBytes []byte) {
	os.MkdirAll(filepath.Dir(Path), os.ModePerm)
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(conf)
	utils.CheckError(err)
	f, err := os.Create(Path)
	utils.CheckError(err)
	defer f.Close()
	f.Write(buf.Bytes())
	appPath := filepath.Join(filepath.Dir(Path), "application.md")
	err = ioutil.WriteFile(appPath, appBytes, 0644)
	utils.CheckError(err)
}

// reloadConf imports the users config onto a default config and then rewrites
// the configuration file.
func reloadConf() {
	result := defaultConf()
	readConf(&result)
	result.General.Version = defaultConf().General.Version
	genConf(result, nil)
}

// defaultApplication defines the default file to be used as an application prompt
// when attempting to submit files.
func defaultApplication() []byte {
	return []byte(
`### Note: any text <!-- inside --> those arrows will be omitted from the submission. Same for lines that start with "#". 
### View this document as raw Markdown instead of rendered HTML to see the prompts.
<!-- Where should your addition be located within the keyset repository?
This line should be in the format of a path.
For example,
library/fiction/classics
or
science/biology/datasets
(An empty line will add the file to the root of the KeySet which is not normally recommended.) -->
# CATEGORY below


<!-- Provide a name for the keyset file that is about to be created (no file extension, just the name) -->
# FILENAME below


<!-- Briefly describe the files you're submitting (preferably <50 characters). -->
# TITLE below


<!-- An empty commit message will abort the submission.
Describe the files in more detail. -->
# COMMIT below


<!-- If you will be submitting a pull request, explain why these files should be added
to the desired repository -->
# PULL REQUEST below

`)
}
