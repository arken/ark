package main

import (
	"ait/cli"
	"fmt"
	"os"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		args = args[1:]
	} else {
		fmt.Println("Usage: ait [command]") //eventually add a real usage printout
		return
	}
	if !cli.IsAITRepo() && args[0] != "init" {
		fmt.Println("this isn't an ait repository. Run \"ait init\"" +
			" before taking further action")
	}
	var err error
	switch args[0] {
	case "add":
		err = cli.Add(args[1:])
	case "status":
		cli.Status()
	case "remove":
		err = cli.Remove(args[1:])
	case "init":
		err = cli.Init()
	}
	if err != nil {
		fmt.Println( err.Error())
	}
}
