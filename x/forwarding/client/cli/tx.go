package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/noble-assets/forwarding/x/forwarding/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  types.ModuleName,
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(TxRegisterAccount())
	cmd.AddCommand(TxRegisterAccountSignerlessly())
	cmd.AddCommand(TxClearAccount())

	return cmd
}

func TxRegisterAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-account [channel] [recipient]",
		Short: "Register a forwarding account for a channel and recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterAccount{
				Signer:    clientCtx.GetFromAddress().String(),
				Recipient: args[1],
				Channel:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func TxRegisterAccountSignerlessly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-account-signerlessly [channel] [recipient]",
		Short: "Signerlessly register a forwarding account for a channel and recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := types.GenerateAddress(args[0], args[1])
			msg := &types.MsgRegisterAccount{
				Signer:    address.String(),
				Recipient: args[1],
				Channel:   args[0],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			factory := tx.NewFactoryCLI(clientCtx, cmd.Flags())
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

func TxClearAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-account [address]",
		Short: "Manually clear funds inside forwarding account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgClearAccount{
				Signer:  clientCtx.GetFromAddress().String(),
				Address: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
