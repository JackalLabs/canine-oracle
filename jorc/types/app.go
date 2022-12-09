package types

import "os"

var (
	NodeDir        = ".jackal-oracle"
	DefaultAppHome = os.ExpandEnv("$HOME/") + NodeDir
)
