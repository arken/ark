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
	switch args[0] {
	case "add":
		err := cli.Add(args[1:])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "status":
		err := cli.Status()
		if err != nil {
			fmt.Println(err)
		}
	}
}
