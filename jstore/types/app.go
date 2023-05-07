package types

import "os"

var (
	NodeDir        = ".jackal-server"
	DefaultAppHome = os.ExpandEnv("$HOME/") + NodeDir
)
