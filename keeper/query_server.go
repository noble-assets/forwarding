// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errorstypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/noble-assets/forwarding/v2/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) Denoms(ctx context.Context, req *types.QueryDenoms) (*types.QueryDenomsResponse, error) {
	if req == nil {
		return nil, errorstypes.ErrInvalidRequest
	}

	allowedDenoms := k.GetAllowedDenoms(ctx)

	return &types.QueryDenomsResponse{AllowedDenoms: allowedDenoms}, nil
}

func (k *Keeper) Address(ctx context.Context, req *types.QueryAddress) (*types.QueryAddressResponse, error) {
	if req == nil {
		return nil, errorstypes.ErrInvalidRequest
	}

	if req.Fallback != "" {
		_, err := k.accountKeeper.AddressCodec().StringToBytes(req.Fallback)
		if err != nil {
			return nil, errors.Wrap(err, "invalid fallback address")
		}
	}

	address := types.GenerateAddress(req.Channel, req.Recipient, req.Fallback)

	exists := false
	if k.accountKeeper.HasAccount(ctx, address) {
		account := k.accountKeeper.GetAccount(ctx, address)
		_, exists = account.(*types.ForwardingAccount)
	}

	return &types.QueryAddressResponse{
		Address: address.String(),
		Exists:  exists,
	}, nil
}

func (k *Keeper) Stats(ctx context.Context, req *types.QueryStats) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, errorstypes.ErrInvalidRequest
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
	if req == nil {
		return nil, errorstypes.ErrInvalidRequest
	}

	numOfAccounts, _ := k.NumOfAccounts.Get(ctx, req.Channel)
	numOfForwards, _ := k.NumOfForwards.Get(ctx, req.Channel)

	return &types.QueryStatsByChannelResponse{
		NumOfAccounts:  numOfAccounts,
		NumOfForwards:  numOfForwards,
		TotalForwarded: k.GetTotalForwarded(ctx, req.Channel),
	}, nil
}
