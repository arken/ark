package keysets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/utils"
	"github.com/schollz/progressbar/v3"
)

const delimiter = "  "

// Generate is the public facing function for the creation of a keyset file.
// Depending on the value of overwrite, the keyset file is either generated from
// scratch or added to.
func Generate(path string, overwrite bool) error {
	ipfs.Init()
	if overwrite {
		return createNew(path)
	}
	return amendExisting(path)
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

	// For large Datasets display a loading bar.
	ipfsBar := progressbar.Default(int64(len(contents)))
	barPresent := false
	if len(contents) > 30 {
		fmt.Println("Adding Files to Embedded IPFS Node:")
		ipfsBar.RenderBlank()
		barPresent = true
	}

	for filePath := range contents {
		line := getKeySetLineFromPath(filePath)
		_, err = keySetFile.WriteString(line + "\n")
		if err != nil {
			cleanup(keySetFile)
			return err
		}
		if barPresent {
			ipfsBar.Add(1)
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

	// For large Datasets display a loading bar.
	namesBar := progressbar.Default(int64(len(newFiles)))
	barPresent := false
	if len(newFiles) > 30 {
		fmt.Println("Reading File Names:")
		namesBar.RenderBlank()
		barPresent = true
	}

	// ^ paths of the files which will be added
	for cid, path := range addedFilesContents {
		if _, contains := ksContents[cid]; !contains {
			filename := strings.Join(strings.Fields(filepath.Base(path)), "-")
			newFiles[cid] = filename
		} else {
			delete(ksContents, cid)
		}
		if barPresent {
			namesBar.Add(1)
		}
	}

	ipfsBar := progressbar.Default(int64(len(newFiles)))
	if barPresent {
		fmt.Println("Adding Files to Embedded IPFS Node:")
		ipfsBar.RenderBlank()
	}

	for cid, filename := range newFiles {
		line := getKeySetLine(filename, cid)
		_, err := keySetFile.WriteString(line + "\n")
		if err != nil {
			return err
		}
		if barPresent {
			ipfsBar.Add(1)
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
func getKeySetLineFromPath(filePath string) string {
	// Scrub filename for spaces and replace with dashes.
	cid, err := ipfs.Add(filePath)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	filename := strings.Join(strings.Fields(filepath.Base(filePath)), "-")
	return getKeySetLine(filename, cid)
}

// getKeySetLine returns a properly formed line for a KeySet file. It expects a
// fileNAME (not path) and an IPFS cid. No newline at the end.
func getKeySetLine(filename, cid string) string {
	// Scrub filename for spaces and replace with dashes.
	filename = strings.Join(strings.Fields(filename), "-")
	return cid + delimiter + filename
}

// fillMapWithCID will fill the given map with IPFS cid hashes as the key and
// either the filename or filepath as the value. This function ONLY be used for
// files that are standard keyset files or files that are just newline separated
// paths. Returns the length of the longest fileNAME, not path. If the file was
// a keyset file, the values are filenames. If the file was just file paths, the
// the values will be file paths.
func fillMapWithCID(contents map[string]string, file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	if filepath.Ext(file.Name()) == ".ks" {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				pair := strings.Fields(line)
				if len(pair) != 2 {
					utils.FatalWithCleanup(utils.SubmissionCleanup,
						"Malformed KeySet file detected:", file.Name())
				}
				contents[pair[0]] = pair[1]
			}
		}
	} else {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) > 0 {
				filename := filepath.Base(line)
				cid, err := ipfs.Add(line)
				utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
				contents[cid] = filename
			}
		}
	}
}
