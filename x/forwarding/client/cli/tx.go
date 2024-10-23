package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transactions commands for the %s module", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(TxRegisterAccountSignerlessly())

	return cmd
}

func TxRegisterAccountSignerlessly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-account-signerlessly [channel] [recipient] (fallback)",
		Short: "Signerlessly register a forwarding account for a channel and recipient",
		Long:  "Signerlessly register a forwarding account for a channel and recipient, with an optional fallback address",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := types.GenerateAddress(args[0], args[1], "")
			if len(args) == 3 {
				address = types.GenerateAddress(args[0], args[1], args[2])
			}
			msg := &types.MsgRegisterAccount{
				Signer:    address.String(),
				Recipient: args[1],
				Channel:   args[0],
			}
			if len(args) == 3 {
				msg.Fallback = args[2]
			}

			factory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			builder, err := factory.BuildUnsignedTx(msg)
			if err != nil {
				return err
			}

			err = builder.SetSignatures(signingtypes.SignatureV2{
				PubKey: &types.ForwardingPubKey{Key: address},
				Data: &signingtypes.SingleSignatureData{
					SignMode:  signingtypes.SignMode_SIGN_MODE_DIRECT,
					Signature: []byte(""),
				},
			})
			if err != nil {
				return err
			}

			if clientCtx.GenerateOnly {
				bz, err := clientCtx.TxConfig.TxJSONEncoder()(builder.GetTx())
				if err != nil {
					return err
				}

				return clientCtx.PrintString(fmt.Sprintf("%s\n", bz))
			}

			bz, err := clientCtx.TxConfig.TxEncoder()(builder.GetTx())
			if err != nil {
				return err
			}
			res, err := clientCtx.BroadcastTx(bz)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
