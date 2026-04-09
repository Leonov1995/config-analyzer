package main

import (
	"os"

	"config-analyzer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(2)
	}
}
