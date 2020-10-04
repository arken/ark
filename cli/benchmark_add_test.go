package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/cmd"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// This test adds every file in your documents folder and reports how long it
// it took. it inits for itself and cleans up after itself.
func BenchmarkBigAdd(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: []string{"."}},
	}
	start := time.Now()
	AddRun(nil, addArgs)
	fmt.Println("\n\t******** Adding all took", time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}

// This test adds every file in your documents folder by adding every folder
// in the Documents folder individually.
func BenchmarkAddManyDirs(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	files, _ := ioutil.ReadDir(testRoot)
	var fileNames []string
	for _, fi := range files {
		fileNames = append(fileNames, fi.Name())
	}
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: fileNames},
	}
	start := time.Now()
	AddRun(nil, addArgs)
	fmt.Println("\n\t******** Adding dirs took", time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}

func BenchmarkAddExtensionFlag(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	ext := "java,c,json,md,js"
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: nil},
		Flags: &AddFlags{Extensions: ext},
	}
	start := time.Now()
	AddRun(nil, addArgs)
	fmt.Println("\n\t******** Adding",ext,"files took", time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}

// This test is for testing performance when adding many individual files.
// Fill them in yourself if you wish to test.
func BenchmarkAddManyFiles(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: []string{
			//put the paths to a bunch of individual files in here for testing
		}},
	}
	start := time.Now()
	AddRun(nil, addArgs)
	fmt.Println("\n\t******** Adding files took", time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}

//func TestUnicode(t *testing.T) {
//	file, err := os.Create("日本語")
//	utils.CheckError(err)
//	info, _ := os.Stat(file.Name())
//	s := info.Name()
//	_ = os.Remove(file.Name())
//	fmt.Println(s)
//	file = utils.BasicFileOpen("utf8test",
//		os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
//	m := map[string]struct{}{ s: {} }
//	err = utils.DumpMap(m, file)
//	utils.CheckError(err)
//	//remember to delete utf8test after checking to make sure s was printed
//	//right and isn't a bunch of gibberish.
//}
