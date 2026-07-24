package main

import (
	"os"

	"github.com/kabroxiko/dayzctl/cmd/dayzctl/commands"
)

func main() {
	// If invoked as `dayzctl rcon ...` run the urfave app (positional UX),
	// otherwise run the legacy Cobra entrypoint so existing commands keep working.
	if len(os.Args) > 1 && os.Args[1] == "rcon" {
		RunApp()
		return
	}
	commands.Execute()
}
