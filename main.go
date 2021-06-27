package main

import (
	"log"
	"os"

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/cmd"
)

func main() {
	if err := cmd.App.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
