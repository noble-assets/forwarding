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

package forwarding

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/forwarding/v2/keeper"
	"github.com/noble-assets/forwarding/v2/types"
)

func InitGenesis(ctx context.Context, k *keeper.Keeper, genesis types.GenesisState) {
	for _, denom := range genesis.AllowedDenoms {
		_ = k.AllowedDenoms.Set(ctx, denom)
	}

	for channel, count := range genesis.NumOfAccounts {
		_ = k.NumOfAccounts.Set(ctx, channel, count)
	}

	for channel, count := range genesis.NumOfForwards {
		_ = k.NumOfForwards.Set(ctx, channel, count)
	}

	for channel, total := range genesis.TotalForwarded {
		_ = k.TotalForwarded.Set(ctx, channel, total)
	}
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		AllowedDenoms:  k.GetAllowedDenoms(ctx),
		NumOfAccounts:  k.GetAllNumOfAccounts(ctx),
		NumOfForwards:  k.GetAllNumOfForwards(ctx),
		TotalForwarded: k.GetAllTotalForwarded(ctx),
	}
}
