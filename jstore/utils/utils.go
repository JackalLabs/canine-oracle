package utils

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
)

func GetDataPath(ctx client.Context) string {
	dataPath := filepath.Join(ctx.HomeDir, "data")

	return dataPath
}
