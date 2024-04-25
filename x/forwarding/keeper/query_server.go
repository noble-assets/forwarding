package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
<<<<<<< HEAD
	"github.com/noble-assets/forwarding/x/forwarding/types"
=======
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) Address(goCtx context.Context, req *types.QueryAddress) (*types.QueryAddressResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	address := types.GenerateAddress(req.Channel, req.Recipient)

	exists := false
	if k.authKeeper.HasAccount(ctx, address) {
		account := k.authKeeper.GetAccount(ctx, address)
		_, exists = account.(*types.ForwardingAccount)
	}

	return &types.QueryAddressResponse{
		Address: address.String(),
		Exists:  exists,
	}, nil
}

<<<<<<< HEAD
func (k *Keeper) StatsByChannel(goCtx context.Context, req *types.QueryStatsByChannel) (*types.QueryStatsByChannelResponse, error) {
=======
func (k *Keeper) Stats(ctx context.Context, req *types.QueryStats) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	stats := make(map[string]types.Stats)

	for channel, numOfAccounts := range k.GetAllNumOfAccounts(ctx) {
		numOfForwards, _ := k.NumOfForwards.Get(ctx, channel)
		totalForwarded := k.GetTotalForwarded(ctx, channel)

		_, clientState, _ := k.channelKeeper.GetChannelClientState(sdk.UnwrapSDKContext(ctx), transfertypes.PortID, channel)

		stats[channel] = types.Stats{
			ChainId:        types.ParseChainId(clientState),
			NumOfAccounts:  numOfAccounts,
			NumOfForwards:  numOfForwards,
			TotalForwarded: totalForwarded,
		}
	}

	return &types.QueryStatsResponse{Stats: stats}, nil
}

func (k *Keeper) StatsByChannel(ctx context.Context, req *types.QueryStatsByChannel) (*types.QueryStatsByChannelResponse, error) {
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryStatsByChannelResponse{
		NumOfAccounts:  k.GetNumOfAccounts(ctx, req.Channel),
		NumOfForwards:  k.GetNumOfForwards(ctx, req.Channel),
		TotalForwarded: k.GetTotalForwarded(ctx, req.Channel),
	}, nil
}
