package server

import (
	"errors"
	"fmt"
	"github.com/JackalLabs/jackal-provider/jorc/crypto"
	"github.com/JackalLabs/jackal-provider/jorc/queue"
	"github.com/JackalLabs/jackal-provider/jorc/utils"
	"github.com/rs/cors"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"os"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
)

func StartFileServer(cmd *cobra.Command) {
	clientCtx, qerr := client.GetClientTxContext(cmd)
	if qerr != nil {
		fmt.Println(qerr)
		return
	}

	address, err := crypto.GetAddress(clientCtx)
	if err != nil {
		fmt.Println(err)
		return
	}

	path := utils.GetDataPath(clientCtx)

	db, dberr := leveldb.OpenFile(path, nil)
	if dberr != nil {
		fmt.Println(dberr)
		return
	}
	router := httprouter.New()

	q := queue.New()

	GetRoutes(cmd, router, db, &q)
	PostRoutes(cmd, router, db, &q)

	handler := cors.Default().Handler(router)

	ctx := utils.GetServerContextFromCmd(cmd)

	go postProofs(cmd, db, &q, ctx)
	go NatCycle(cmd.Context())
	go q.StartListener(cmd)
	go q.CheckStrays(cmd, db)

	port, err := cmd.Flags().GetString("port")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("🌍 Started Provider: http://0.0.0.0:%s\n", port)
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), handler)
	if err != nil {
		fmt.Println(err)
		return
	}

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Storage Provider Closed\n")
		return
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
