package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"time"

	provider "github.com/JackalLabs/jackal-provider/jprov/types"
	"github.com/TheMarstonConnell/DelphiHack/server/jstore/crypto"
	oracletypes "github.com/jackalLabs/canine-chain/x/oracle/types"
	"github.com/jackalLabs/canine-chain/x/storage/types"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"github.com/TheMarstonConnell/DelphiHack/server/jstore/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/spf13/cobra"
)

const MaxFileSize = 32 << 30

func fileUpload(w *http.ResponseWriter, r *http.Request, cmd *cobra.Command, db *leveldb.DB) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return
	}
	ctx := utils.GetServerContextFromCmd(cmd)
	address, err := crypto.GetAddress(clientCtx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ParseMultipartForm parses a request body as multipart/form-data
	err = r.ParseMultipartForm(MaxFileSize) // MAX file size lives here
	if err != nil {
		ctx.Logger.Error("Error with parsing form!")
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	file, h, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		ctx.Logger.Error("Error with form file!")
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	writer.WriteField("sender", address)

	fileWriter, err := writer.CreateFormFile("file", h.Filename)
	if err != nil {
		return
	}
	// copy the file into the fileWriter
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	writer.Close()

	req, err := http.NewRequest("POST", "https://jackalplswork.com/upload", &b)
	if err != nil {
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	cli := &http.Client{Timeout: time.Second * 100}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := cli.Do(req)
	if err != nil {
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}
	// Check the response
	if res.StatusCode != http.StatusOK {
		var errRes provider.ErrorResponse
		_ = json.NewDecoder(res.Body).Decode(&errRes)

		err = fmt.Errorf("bad status: %s", errRes.Error)
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	var pup provider.UploadResponse
	err = json.NewDecoder(res.Body).Decode(&pup)
	if err != nil {
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	msg := types.NewMsgSignContract(
		address,
		pup.CID,
		true,
	)

	txRes, err := utils.SendTx(clientCtx, cmd.Flags(), msg)
	if err != nil {
		v := ErrorResponse{
			Error: err.Error(),
		}
		(*w).WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(*w).Encode(v)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
		return
	}

	fmt.Println(txRes)

	nv := UploadResponse{
		FID: pup.FID,
	}
	err = json.NewEncoder(*w).Encode(nv)
	if err != nil {
		ctx.Logger.Error(err.Error())
	}

}

func RunServer(db *leveldb.DB, ctx *client.Context, cmd *cobra.Command) {
	router := httprouter.New()
	handler := cors.Default().Handler(router)

	router.POST("/upload", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fileUpload(&w, r, cmd, db)
	})

	fmt.Printf("🌍 Started Provider: http://0.0.0.0:%d\n", 2929)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", 2929), handler)
	if err != nil {
		fmt.Println(err)
		return
	}

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server Closed\n")
		return
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func StartServer(cmd *cobra.Command) {
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

	fmt.Println("Started Jackal Server 📁")

	go RunServer(db, &ctx, cmd)

	runtime.Goexit()

	fmt.Println("Quit Server")
}

func PrepareOracleData(r *http.Response) ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		return nil, err
	}

	jsonMap := make(map[string](interface{}))
	err = json.Unmarshal([]byte(b), &jsonMap)
	if err != nil {
		return nil, fmt.Errorf("ERROR: fail to unmarshal json, %s\n", err.Error())
	}

	stringMap := make(map[string](string))
	for k, v := range jsonMap {
		stringMap[k] = fmt.Sprint(v)
	}

	m, err := json.Marshal(stringMap)
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
