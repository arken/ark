package keysets

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	err := Generate("test.ks")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	os.Remove("test.ks")
}
