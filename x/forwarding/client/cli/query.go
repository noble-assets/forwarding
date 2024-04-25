package cli

import (
	"context"
<<<<<<< HEAD

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/noble-assets/forwarding/x/forwarding/types"
=======
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/gogoproto/proto"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
<<<<<<< HEAD
		Use:  types.ModuleName,
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(QueryAddress())
	cmd.AddCommand(QueryStats())
=======
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryStats())
>>>>>>> 8ab8bfa (feat: add general stats query (#5))

	return cmd
}

<<<<<<< HEAD
func QueryAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address [channel] [recipient]",
		Short: "Query forwarding address by channel and recipient",
		Args:  cobra.ExactArgs(2),
=======
func CmdQueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats (channel)",
		Short: "Query forwarding stats",
		Args:  cobra.MaximumNArgs(1),
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

<<<<<<< HEAD
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
		Use:   "stats [channel]",
		Short: "Query forwarding stats by channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryStatsByChannel{Channel: args[0]}

			res, err := queryClient.StatsByChannel(context.Background(), req)
=======
			var res proto.Message
			var err error

			if len(args) == 1 {
				res, err = queryClient.StatsByChannel(context.Background(), &types.QueryStatsByChannel{Channel: args[0]})
			} else {
				res, err = queryClient.Stats(context.Background(), &types.QueryStats{})
			}

>>>>>>> 8ab8bfa (feat: add general stats query (#5))
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
