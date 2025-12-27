package main

import (
	"github.com/jidckii/kolor-keyboard/cmd/kolor-keyboard/cmd"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Execute()
}
