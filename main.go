package main

import (
	"os"

	"github.com/anthropics/opencc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
