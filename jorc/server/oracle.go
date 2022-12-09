package server

import (
	"fmt"
	"runtime"
	"time"

	"github.com/JackalLabs/jackal-oracle/jorc/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/spf13/cobra"
)

func RunOracle(db *leveldb.DB, ctx *utils.Context) {
	for {
		fmt.Println("Oracle!")

		time.Sleep(time.Second * 10)
	}
}

func StartOracle(cmd *cobra.Command) {
	clientCtx, qerr := client.GetClientTxContext(cmd)
	if qerr != nil {
		fmt.Println(qerr)
		return
	}

	path := utils.GetDataPath(clientCtx)

	db, dberr := leveldb.OpenFile(path, nil)
	if dberr != nil {
		fmt.Println(dberr)
		return
	}

	ctx := utils.GetServerContextFromCmd(cmd)

	go RunOracle(db, ctx)

	runtime.Goexit()

	fmt.Println("Quit Oracle")
}
