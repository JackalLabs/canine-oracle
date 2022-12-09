package main

import (
	"os"

	"github.com/JackalLabs/jackal-provider/jorc/types"
)

func main() {
	rootCmd := NewRootCmd()

	if err := Execute(rootCmd, types.DefaultAppHome); err != nil {
		os.Exit(1)
	}
}
