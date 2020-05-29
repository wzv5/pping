package main

import (
	"log"

	"github.com/wzv5/pping/cmd/pping/cmd"
)

func main() {
	log.SetFlags(log.Ltime)
	cmd.Execute()
}
