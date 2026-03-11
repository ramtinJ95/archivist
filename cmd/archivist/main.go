package main

import (
	"os"

	"github.com/ramtinJ95/archivist/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
