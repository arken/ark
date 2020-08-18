package keysets

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	err := Generate("test.ks", false)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	os.Remove("test.ks")
}
