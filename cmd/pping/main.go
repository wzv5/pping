package main

import (
	"log"
	"os"

	"github.com/wzv5/pping/cmd/pping/cmd"
)

var version string

func main() {
	log.SetFlags(log.Ltime)
	cmd.Version = version
	err := cmd.Execute()
	if err == cmd.ErrPing {
		os.Exit(1)
	} else if err != nil {
		os.Exit(2)
	}
}
