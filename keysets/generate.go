package keysets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/utils"
)

//Generate creates a keyset file with the given path. Path should not be the
//desired directory, rather it should be a full path to a file which does not
//exist yet, and the file should end in ".ks" The resultant keyset files contains
//the name (not path) of the file and an IPFS cid hash, separated by a space.
func Generate(path string) error {
	keySetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	addedFiles, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	contents := make(map[string]struct{})
	utils.FillMap(contents, addedFiles)
	addedFiles.Close()
	for filePath := range contents {
		cid, err := ipfs.Add(filePath)
		if err != nil {
			cleanup(keySetFile)
			return err
		}
		line := fmt.Sprintf("%v %v\n", filepath.Base(filePath), cid)
		_, err = keySetFile.WriteString(line)
		if err != nil {
			cleanup(keySetFile)
			return err
		}
	}
	return keySetFile.Close()
}

func cleanup(file *os.File) {
	path := file.Name()
	file.Close()
	_ = os.Remove(path)
}
