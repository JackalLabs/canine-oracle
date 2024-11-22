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

		apiLink, err := db.Get([]byte("api"), nil)
		if err != nil {
			fmt.Println(err)
			fmt.Println("No API link, have you ran `jorcd set-feed` yet.")
			return
		}

		name, err := db.Get([]byte("name"), nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		r, err := http.Get(string(apiLink))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Missed a GET, will try again.")
			continue
		}

		defer r.Body.Close()

		data, err := PrepareOracleData(r)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Posting to %s: %s\n", name, string(data))

		SendMsgUpdateFeed(ctx, cmd, name, data)

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
	fmt.Println("---->", path)

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

type ChainExpects struct {
	Price  string `json:"price"`
	Change string `json:"24h_change"`
}

type CoinGecko struct {
	Jackal JackalPrice `json:"jackal-protocol"`
}

type JackalPrice struct {
	USD float64 `json:"usd"`
}

func PrepareOracleData(r *http.Response) ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		return nil, err
	}

	var jackal CoinGecko
	err = json.Unmarshal(b, &jackal)
	if err != nil {
		return nil, fmt.Errorf("ERROR: fail to unmarshal json, %s\n", err.Error())
	}

	ce := ChainExpects{
		Change: "0.0",
		Price:  fmt.Sprintf("%f", jackal.Jackal.USD),
	}

	m, err := json.Marshal(ce)
	if err != nil {
		return nil, fmt.Errorf("ERROR: fail to marshal json, %s\n", err.Error())
	}

	return m, nil
}

func SendMsgUpdateFeed(ctx *client.Context, cmd *cobra.Command, name []byte, data []byte) {
	address, err := crypto.GetAddress(*ctx)
	if err != nil {
		panic(err)
	}

	msg := oracletypes.NewMsgUpdateFeed(
		address,
		string(name),
		string(data),
	)
	if err := msg.ValidateBasic(); err != nil {
		fmt.Println(err)
		return
	}

	_, err = utils.SendTx(*ctx, cmd.Flags(), msg)
	if err != nil {
		fmt.Println(err)
	}
}
