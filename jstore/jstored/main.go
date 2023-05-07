package main

import (
	"os"

	"github.com/TheMarstonConnell/DelphiHack/server/jstore/types"
)

func main() {
	rootCmd := NewRootCmd()

	if err := Execute(rootCmd, types.DefaultAppHome); err != nil {
		os.Exit(1)
	}
}
