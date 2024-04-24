package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) Address(ctx context.Context, req *types.QueryAddress) (*types.QueryAddressResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

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

func (k *Keeper) StatsByChannel(ctx context.Context, req *types.QueryStatsByChannel) (*types.QueryStatsByChannelResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	numOfAccounts, _ := k.NumOfAccounts.Get(ctx, req.Channel)
	numOfForwards, _ := k.NumOfForwards.Get(ctx, req.Channel)

	return &types.QueryStatsByChannelResponse{
		NumOfAccounts:  numOfAccounts,
		NumOfForwards:  numOfForwards,
		TotalForwarded: k.GetTotalForwarded(ctx, req.Channel),
	}, nil
}
