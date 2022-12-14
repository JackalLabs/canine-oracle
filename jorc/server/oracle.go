package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/JackalLabs/jackal-oracle/jorc/crypto"
	oracletypes "github.com/jackalLabs/canine-chain/x/oracle/types"

	"github.com/JackalLabs/jackal-oracle/jorc/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/spf13/cobra"
)

func RunOracle(db *leveldb.DB, ctx *client.Context, cmd *cobra.Command) {
	interval, err := db.Get([]byte("interval"), nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("You most likely haven't ran `jorcd set-feed` yet.")
		return
	}

	t, err := time.ParseDuration(fmt.Sprintf("%ss", string(interval)))
	if err != nil {
		panic(err)
	}

	for {
		address, err := crypto.GetAddress(*ctx)
		if err != nil {
			panic(err)
		}

		res, err := db.Get([]byte("api"), nil)
		if err != nil {
			fmt.Println(err)
			fmt.Println("You most likely haven't ran `jorcd set-feed` yet.")
			return
		}

		name, err := db.Get([]byte("name"), nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		r, err := http.Get(string(res))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Missed a GET, will try again.")
			continue
		}

		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
		if err != nil {
			fmt.Println(err)
		}

		jsonMap := make(map[string](interface{}))
		err = json.Unmarshal([]byte(b), &jsonMap)
		if err != nil {
			fmt.Printf("ERROR: fail to unmarshal json, %s\n", err.Error())
		}

		stringMap := make(map[string](string))
		for k, v := range jsonMap {
			stringMap[k] = fmt.Sprint(v)
		}

		m, err := json.Marshal(stringMap)
		if err != nil {
			fmt.Printf("ERROR: fail to marshal json, %s\n", err.Error())
		}

		fmt.Printf("Posting to %s: %s\n", name, string(m))

		msg := oracletypes.NewMsgUpdateFeed(
			address,
			string(name),
			string(m),
		)
		if err := msg.ValidateBasic(); err != nil {
			fmt.Println(err)
			return
		}

		_, err = utils.SendTx(*ctx, cmd.Flags(), msg)
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(t)
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

	ctx, err := client.GetClientTxContext(cmd)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Started Jackal Oracle 👁️")

	go RunOracle(db, &ctx, cmd)

	runtime.Goexit()

	fmt.Println("Quit Oracle")
}
