package main

import (
	"bufio"
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/version"

	"github.com/JackalLabs/jackal-oracle/jorc/crypto"
	"github.com/JackalLabs/jackal-oracle/jorc/server"
	"github.com/cosmos/cosmos-sdk/client"
	clientConfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stortypes "github.com/jackalLabs/canine-chain/x/storage/types"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

func StartServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start jackal oracle",
		Long:  `Start the Jackal Oracle with the current settings.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			server.StartOracle(cmd)
			return nil
		},
	}

	AddTxFlagsToCmd(cmd)
	return cmd
}

func InitOracleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init jackal oracle",
		Long:  `Initialize the Jackal Oracle with new settings.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			server.StartOracle(cmd)
			return nil
		},
	}

	AddTxFlagsToCmd(cmd)
	return cmd
}

func NetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Oracle network commands",
		Long:  `The sub-menu for Jackal Oracle network commands.`,
	}

	cmd.AddCommand(
		rpc.StatusCommand(),
	)

	return cmd
}

func ClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Oracle client commands",
		Long:  `The sub-menu for Jackal Oracle client commands.`,
	}

	cmd.AddCommand(
		GenKeyCommand(),
		clientConfig.Cmd(),
		GetBalanceCmd(),
		GetAddressCmd(),
	)

	return cmd
}

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version info",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s %s\n", version.AppName, version.Version)
			return nil
		},
	}

	return cmd
}

func GetBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get account balance",
		Long:  `Get the account balance of the current oracle key.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := banktypes.NewQueryClient(clientCtx)
			address, err := crypto.GetAddress(clientCtx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			params := &banktypes.QueryBalanceRequest{
				Denom:   "ujkl",
				Address: address,
			}

			res, err := queryClient.Balance(context.Background(), params)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("Balance: %s\n", res.Balance)

			return nil
		},
	}

	return cmd
}

func GetAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Get account address",
		Long:  `Get the account address of the current oracle key.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, err := crypto.GetAddress(clientCtx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Printf("Address: %s\n", address)

			return nil
		},
	}

	return cmd
}

func GenKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-key",
		Short: "Generate a new private key",
		Long:  `Generate a new Jackal address and private key combination to interact with the blockchain.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			pKey := secp256k1.GenPrivKey()
			address, err := bech32.ConvertAndEncode(stortypes.AddressPrefix, pKey.PubKey().Address().Bytes())
			if err != nil {
				return err
			}

			keyExport := crypto.StorPrivKey{
				Address: address,
				Key:     crypto.ExportPrivKey(pKey),
			}

			alreadyKey := crypto.KeyExists(clientCtx)
			buf := bufio.NewReader(cmd.InOrStdin())

			if alreadyKey {
				yes, err := input.GetConfirmation("Key already exists, would you like to overwrite it? If so please make sure you have created a backup.", buf, cmd.ErrOrStderr())
				if err != nil {
					return err
				}

				if !yes {
					return nil
				}
			}

			err = crypto.WriteKey(clientCtx, &keyExport)
			if err != nil {
				return err
			}

			fmt.Printf("Your new address is %s\n", address)

			return nil
		},
	}

	AddTxFlagsToCmd(cmd)

	return cmd
}

// AddTxFlagsToCmd adds common flags to a module tx command.
func AddTxFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "json", "Output format (text|json)")
	cmd.Flags().Uint64P(flags.FlagAccountNumber, "a", 0, "The account number of the signing account (offline mode only)")
	cmd.Flags().Uint64P(flags.FlagSequence, "s", 0, "The sequence number of the signing account (offline mode only)")
	cmd.Flags().String(flags.FlagNote, "", "Note to add a description to the transaction (previously --memo)")
	cmd.Flags().String(flags.FlagFees, "", "Fees to pay along with transaction; eg: 10ujkl")
	cmd.Flags().String(flags.FlagGasPrices, "0.02ujkl", "Gas prices in decimal format to determine the transaction fee (e.g. 0.1ujkl)")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().Bool(flags.FlagUseLedger, false, "Use a connected Ledger device")
	cmd.Flags().Float64(flags.FlagGasAdjustment, flags.DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
	cmd.Flags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastBlock, "Transaction broadcasting mode (sync|async|block)")
	cmd.Flags().Bool(flags.FlagDryRun, false, "ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)")
	cmd.Flags().Bool(flags.FlagGenerateOnly, false, "Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible)")
	cmd.Flags().Bool(flags.FlagOffline, false, "Offline mode (does not allow any online functionality")
	cmd.Flags().BoolP(flags.FlagSkipConfirmation, "y", false, "Skip tx broadcasting prompt confirmation")
	cmd.Flags().String(flags.FlagSignMode, "", "Choose sign mode (direct|amino-json), this is an advanced feature")
	cmd.Flags().Uint64(flags.FlagTimeoutHeight, 0, "Set a block timeout height to prevent the tx from being committed past a certain height")
	cmd.Flags().String(flags.FlagFeeAccount, "", "Fee account pays fees for the transaction instead of deducting from the signer")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")

	// --gas can accept integers and "auto"
	cmd.Flags().String(flags.FlagGas, "", fmt.Sprintf("gas limit to set per-transaction; set to %q to calculate sufficient gas automatically (default %d)", flags.GasFlagAuto, flags.DefaultGasLimit))
}
