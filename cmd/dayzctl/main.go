package main

import "github.com/kabroxiko/dayzctl/cmd/dayzctl/commands"

func main() {
	// Use Cobra's Execute() to ensure persistent flags and PersistentPreRunE
	// are parsed/executed (loads config, initializes logger, etc.).
	commands.Execute()
}
