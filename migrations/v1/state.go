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

package v1

import (
	"strconv"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/noble-assets/forwarding/v2/types"
)

// GetAllNumOfAccounts implements adapted legacy store logic from version 1.
// https://github.com/noble-assets/forwarding/blob/v1.x/x/forwarding/keeper/state.go#L25-L39
func GetAllNumOfAccounts(adapter storetypes.KVStore) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(adapter, types.NumOfAccountsPrefix)
	iterator := store.Iterator(nil, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}

// GetAllNumOfForwards implements adapted legacy store logic from version 1.
// https://github.com/noble-assets/forwarding/blob/v1.x/x/forwarding/keeper/state.go#L71-L85
func GetAllNumOfForwards(adapter storetypes.KVStore) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(adapter, types.NumOfForwardsPrefix)
	iterator := store.Iterator(nil, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}
