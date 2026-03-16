package main

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
