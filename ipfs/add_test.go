package ipfs

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	err := ioutil.WriteFile("test", []byte("hello world"), os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	cid, err := Add("test")
	if err != nil {
		t.Error(err)
	}
	if cid != "Qmf412jQZiuVUtdgnB36FXFX7xg5V6KEbSJ4dpQuhkLyfD" {
		t.Errorf("Wrong CID")
	}

	err = os.Remove("test")
	if err != nil {
		t.Error(err)
	}
}
