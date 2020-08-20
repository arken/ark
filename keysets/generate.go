package keysets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/utils"
)

// Generate is the public facing function for the creation of a keyset file.
// Depending on the value of overwrite, the keyset file is either generated from
// scratch or added to.
func Generate(path string, overwrite bool) error {
	if overwrite {
		return createNew(path)
	} else {
		return amendExisting(path)
	}
}

// createNew creates a keyset file with the given path. Path should not be the
// desired directory, rather it should be a full path to a file which does not
// exist yet (will be truncated if it does exist), and the file should end in
// ".ks" The resultant keyset files contains the name (not path) of the file and
// an IPFS cid hash, separated by a space.
func createNew(path string) error {
	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)

	keySetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
	var maxTitle int
	for filePath := range contents {
		filename := strings.Join(strings.Fields(filepath.Base(filePath)), "-")
		maxTitle = utils.IMax(maxTitle, len(filename))
	}
	for filePath := range contents {
		line := getKeySetLineFromPath(filePath, maxTitle)
		_, err = keySetFile.WriteString(line + "\n")
		if err != nil {
			cleanup(keySetFile)
			return err
		}
	}
	return keySetFile.Close()
}

// amendExisting looks at current files in added_files and adds any that aren't
// already in the keyset file to the keyset files. The keyset file in question
// should be at path.
func amendExisting(ksPath string) error {
	keySetFile, err := os.OpenFile(ksPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer keySetFile.Close()
	addedFiles, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer addedFiles.Close()
	addedFilesContents := make(map[string]string)
	// ^ cid -> filePATH
	fillMapWithCID(addedFilesContents, addedFiles)
	ksContents := make(map[string]string)
	// ^ cid -> fileNAME
	fillMapWithCID(ksContents, keySetFile)
	newFiles := make(map[string]string)
	// ^ paths of the files which will be added
	max := 0
	for cid, path := range addedFilesContents {
		if _, contains := ksContents[cid]; !contains {
			filename := strings.Join(strings.Fields(filepath.Base(path)), "-")
			newFiles[cid] = filename
			max = utils.IMax(max, len(filename))
		} else {
			delete(ksContents, cid)
		}
	}
	for cid, filename := range newFiles {
		line := getKeySetLine(filename, cid, max)
		_, err := keySetFile.WriteString(line+"\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// cleanup closes and deletes the given file.
func cleanup(file *os.File) {
	path := file.Name()
	file.Close()
	_ = os.Remove(path)
}

// getKeySetLine returns a properly formed line for a KeySet file given a path
// to a file. No newline at the end.
func getKeySetLineFromPath(filePath string, maxTitle int) string {
	// Scrub filename for spaces and replace with dashes.
	cid, err := ipfs.Add(filePath)
	utils.CheckError(err)
	filename := strings.Join(strings.Fields(filepath.Base(filePath)), "-")
	return fmt.Sprintf("%v%v%v", filename,
		strings.Repeat(" ", maxTitle+4-len(filename)), cid)
}

// getKeySetLine returns a properly formed line for a KeySet file. It expects a
// fileNAME (not path) and an IPFS cid. No newline at the end.
func getKeySetLine(filename, cid string, maxTitle int) string {
	// Scrub filename for spaces and replace with dashes.
	filename = strings.Join(strings.Fields(filename), "-")
	return fmt.Sprintf("%v%v%v", filename,
		strings.Repeat(" ", maxTitle+4-len(filename)), cid)
}

// fillMapWithCID will fill the given map with IPFS cid hashes as the key and
// either the filename or filepath as the value. This function ONLY be used for
// files that are standard keyset files or files that are just newline separated
// paths. Returns the length of the longest fileNAME, not path. If the file was
// a keyset file, the values are filenames. If the file was just file paths, the
// the values will be file paths.
func fillMapWithCID(contents map[string]string, file *os.File) int {
	max := 0
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	if filepath.Ext(file.Name()) == ".ks" {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				pair := strings.Fields(line)
				if len(pair) != 2 {
					utils.FatalPrintln("Malformed KeySet file detected.")
				}
				max = utils.IMax(max, len(pair[0]))
				contents[pair[1]] = pair[0]
			}
		}
	} else {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				filename := filepath.Base(line)
				max = utils.IMax(max, len(filename))
				cid, err := ipfs.Add(line)
				utils.CheckError(err)
				contents[cid] = filename
			}
		}
	}
	return max
}
