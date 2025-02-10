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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/forwarding/v2/types"
)

// PERSISTENT STATE

func (k *Keeper) GetAllowedDenoms(ctx context.Context) []string {
	var denoms []string

	_ = k.AllowedDenoms.Walk(ctx, nil, func(denom string) (stop bool, err error) {
		denoms = append(denoms, denom)
		return false, nil
	})

	return denoms
}

func (k *Keeper) GetAllNumOfAccounts(ctx context.Context) map[string]uint64 {
	counts := make(map[string]uint64)

	_ = k.NumOfAccounts.Walk(ctx, nil, func(key string, value uint64) (stop bool, err error) {
		counts[key] = value

		return false, nil
	})

	return counts
}

func (k *Keeper) GetAllNumOfForwards(ctx context.Context) map[string]uint64 {
	counts := make(map[string]uint64)

	_ = k.NumOfForwards.Walk(ctx, nil, func(key string, value uint64) (stop bool, err error) {
		counts[key] = value

		return false, nil
	})

	return counts
}

func (k *Keeper) IncrementNumOfAccounts(ctx context.Context, channel string) {
	count, _ := k.NumOfAccounts.Get(ctx, channel)
	_ = k.NumOfAccounts.Set(ctx, channel, count+1)

	k.Logger().Info("registered a new account", "channel", channel)
}

func (k *Keeper) IncrementNumOfForwards(ctx context.Context, channel string) {
	count, _ := k.NumOfForwards.Get(ctx, channel)
	_ = k.NumOfForwards.Set(ctx, channel, count+1)
}

func (k *Keeper) GetTotalForwarded(ctx context.Context, channel string) sdk.Coins {
	rawTotal, _ := k.TotalForwarded.Get(ctx, channel)
	total, _ := sdk.ParseCoinsNormalized(rawTotal)
	return total
}

func (k *Keeper) GetAllTotalForwarded(ctx context.Context) map[string]string {
	totals := make(map[string]string)

	_ = k.TotalForwarded.Walk(ctx, nil, func(key string, value string) (stop bool, err error) {
		totals[key] = value

		return false, nil
	})

	return totals
}

func (k *Keeper) IncrementTotalForwarded(ctx context.Context, channel string, coin sdk.Coin) {
	total := k.GetTotalForwarded(ctx, channel)
	_ = k.TotalForwarded.Set(ctx, channel, total.Add(coin).String())
}

// TRANSIENT STATE

func (k *Keeper) GetPendingForwards(ctx context.Context) (accounts []types.ForwardingAccount) {
	_ = k.PendingForwards.Walk(ctx, nil, func(key string, value types.ForwardingAccount) (stop bool, err error) {
		accounts = append(accounts, value)

		return false, nil
	})

	return
}

func (k *Keeper) SetPendingForward(ctx context.Context, account *types.ForwardingAccount) {
	if found, err := k.PendingForwards.Has(ctx, account.Address); err != nil || found {
		return
	}

	_ = k.PendingForwards.Set(ctx, account.Address, *account)
}
