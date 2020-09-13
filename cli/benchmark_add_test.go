package cli

import (
	"github.com/DataDrake/cli-ng/cmd"
	"os"
	"path/filepath"
	"testing"
)

func TestBigAdd(t *testing.T) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil) //args are never used in InitRun, this is fine
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: []string{"."}},
	}
	AddRun(nil, addArgs)
	_ = os.RemoveAll(testRoot + "/.ait")
}

func TestAddManyDirs(t *testing.T) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	addArgs := &cmd.CMD{
		//put any or all of the folders in your documents in this slice for testing
		Args: &AddArgs{Paths: []string{}},
	}
	AddRun(nil, addArgs)
	_ = os.RemoveAll(testRoot + "/.ait")
}

func TestAddManyFiles(t *testing.T) {
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
	AddRun(nil, addArgs)
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
