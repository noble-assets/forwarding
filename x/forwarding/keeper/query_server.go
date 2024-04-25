package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/noble-assets/forwarding/x/forwarding/types"
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

func (k *Keeper) Stats(goCtx context.Context, req *types.QueryStats) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	stats := make(map[string]types.Stats)

	for channel, numOfAccounts := range k.GetAllNumOfAccounts(ctx) {
		numOfForwards := k.GetNumOfForwards(ctx, channel)
		totalForwarded := k.GetTotalForwarded(ctx, channel)

		_, clientState, _ := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, channel)

		stats[channel] = types.Stats{
			ChainId:        types.ParseChainId(clientState),
			NumOfAccounts:  numOfAccounts,
			NumOfForwards:  numOfForwards,
			TotalForwarded: totalForwarded,
		}
	}

	return &types.QueryStatsResponse{Stats: stats}, nil
}

func (k *Keeper) StatsByChannel(goCtx context.Context, req *types.QueryStatsByChannel) (*types.QueryStatsByChannelResponse, error) {
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
