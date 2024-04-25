package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/gogo/protobuf/proto"
	"github.com/noble-assets/forwarding/x/forwarding/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(QueryAddress())
	cmd.AddCommand(QueryStats())

	return cmd
}

func QueryAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address [channel] [recipient]",
		Short: "Query forwarding address by channel and recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAddress{Channel: args[0], Recipient: args[1]}

			res, err := queryClient.Address(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats (channel)",
		Short: "Query forwarding stats",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			var res proto.Message
			var err error

			if len(args) == 1 {
				res, err = queryClient.StatsByChannel(context.Background(), &types.QueryStatsByChannel{Channel: args[0]})
			} else {
				res, err = queryClient.Stats(context.Background(), &types.QueryStats{})
			}

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
