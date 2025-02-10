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

package types

import (
	"bytes"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

var (
	_ sdk.AccountI             = &ForwardingAccount{}
	_ authtypes.GenesisAccount = &ForwardingAccount{}
)

func GenerateAddress(channel string, recipient string, fallback string) sdk.AccAddress {
	bz := []byte(channel + recipient + fallback)
	return address.Derive([]byte(ModuleName), bz)[12:]
}

func (fa *ForwardingAccount) Validate() error {
	if !channeltypes.IsValidChannelID(fa.Channel) {
		return fmt.Errorf("%s is an invalid channel id", fa.Channel)
	}

	if fa.CreatedAt < 0 {
		return fmt.Errorf("%d is an invalid creation block height", fa.CreatedAt)
	}

	return fa.BaseAccount.Validate()
}

//

var _ cryptotypes.PubKey = &ForwardingPubKey{}

func (fpk *ForwardingPubKey) String() string {
	return fmt.Sprintf("PubKeyForwarding{%X}", fpk.Key)
}

func (fpk *ForwardingPubKey) Address() cryptotypes.Address { return fpk.Key }

func (fpk *ForwardingPubKey) Bytes() []byte { return fpk.Key }

func (*ForwardingPubKey) VerifySignature(_ []byte, _ []byte) bool {
	panic("PubKeyForwarding.VerifySignature should never be invoked")
}

func (fpk *ForwardingPubKey) Equals(other cryptotypes.PubKey) bool {
	if _, ok := other.(*ForwardingPubKey); !ok {
		return false
	}

	return bytes.Equal(fpk.Bytes(), other.Bytes())
}

func (*ForwardingPubKey) Type() string { return "forwarding" }
