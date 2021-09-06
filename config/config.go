package config

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	// Version is the current version of Arken
	Version string = "develop"
	// Global is the global application configuration
	Global Config
)

type Config struct {
	Core     core     `toml:"core"`
	Manifest manifest `toml:"manifest"`
	Git      git      `toml:"git"`
}

type core struct {
	Editor string `toml:"editor"`
}

type git struct {
	Name     string `toml:"name"`
	Username string `toml:"username"`
	Email    string `toml:"email"`
	Token    string `toml:"token"`
}

type manifest struct {
	Path    string            `toml:"path"`
	Aliases map[string]string `toml:"aliases"`
}

func Init(path string) error {
	// Generate the default config
	Global = Config{
		Core: core{
			Editor: "nano",
		},
		Manifest: manifest{
			Path:    filepath.Join(filepath.Dir(path), "manifest"),
			Aliases: make(map[string]string),
		},
		Git: git{
			Name:  "",
			Email: "",
			Token: "",
		},
	}

	// Setup default alias for core-manifest
	Global.Manifest.Aliases["core"] = "https://github.com/arken/core-manifest"

	// Read in config from file
	err := ParseFile(path, &Global)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Read in config from environment
	err = sourceEnv(&Global)
	if err != nil {
		return err
	}

	// Write config file
	err = WriteFile(path, &Global)
	return err
}

// ParseFile decodes the application configuration
// from the TOML encoded file at the specified path.
func ParseFile(path string, in *Config) error {
	_, err := toml.DecodeFile(path, in)
	return err
}

func sourceEnv(in *Config) error {
	numSubStructs := reflect.ValueOf(in).Elem().NumField()
	// Check for env args matching each of the sub structs.
	for i := 0; i < numSubStructs; i++ {
		iter := reflect.ValueOf(in).Elem().Field(i)
		subStruct := strings.ToUpper(iter.Type().Name())
		structType := iter.Type()
		for j := 0; j < iter.NumField(); j++ {
			fieldVal := iter.Field(j).String()
			fieldName := structType.Field(j).Name
			evName := "ARK" + "_" + subStruct + "_" + strings.ToUpper(fieldName)
			evVal, evExists := os.LookupEnv(evName)
			if evExists && evVal != fieldVal {
				iter.FieldByName(fieldName).SetString(evVal)
			}
		}
	}
	return nil
}

// WriteFile writes changes to the application configuration back
// to the TOML encoded file.
func WriteFile(path string, in *Config) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(in)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, buf.Bytes(), os.ModePerm)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, buf.Bytes(), os.ModePerm)
	}
	return err
}
