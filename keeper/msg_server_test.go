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
package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/noble-assets/forwarding/v2/keeper"
	"github.com/stretchr/testify/require"

	"github.com/noble-assets/forwarding/v2/types"
)

func TestValidateAccountFields(t *testing.T) {
	key := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(key.PubKey().Address())

	tests := []struct {
		name        string
		malleate    func(acc sdk.AccountI)
		errContains string
	}{
		{
			name:     "New account",
			malleate: func(acc sdk.AccountI) {},
		},
		{
			name:        "Account with nil pub key but non zero sequence",
			malleate:    func(acc sdk.AccountI) { acc.SetSequence(1) },
			errContains: "attempting to register an existing user",
		},
		{
			name: "Account created signerlessly",
			malleate: func(acc sdk.AccountI) {
				acc.SetPubKey(&types.ForwardingPubKey{Key: addr})
			},
		},
		{
			name: "Account created signerlessly and non zero sequence",
			malleate: func(acc sdk.AccountI) {
				acc.SetPubKey(&types.ForwardingPubKey{Key: addr})
				acc.SetSequence(100)
			},
		},
		{
			name: "Account created signerlessly but wrong address",
			malleate: func(acc sdk.AccountI) {
				key := secp256k1.GenPrivKey()
				newAddr := sdk.AccAddress(key.PubKey().Address())
				acc.SetPubKey(&types.ForwardingPubKey{Key: newAddr})
				acc.SetAddress(newAddr)
			},
			errContains: "attempting to register an existing user",
		},
		{
			name: "Account created with different pub key type",
			malleate: func(acc sdk.AccountI) {
				key := secp256k1.GenPrivKey()
				acc.SetPubKey(key.PubKey())
			},
			errContains: "attempting to register an existing user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseAcc := &authtypes.BaseAccount{Address: addr.String()}
			tt.malleate(baseAcc)

			err := keeper.ValidateAccountFields(baseAcc, sdk.AccAddress(addr))
			if tt.errContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
