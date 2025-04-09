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
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/noble-assets/forwarding/v2/types"
)

// SigVerificationGasConsumer is a wrapper around the default provided by the
// Cosmos SDK that supports forwarding account public keys.
func SigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	switch sig.PubKey.(type) {
	case *types.ForwardingPubKey:
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}

//

var _ sdk.AnteDecorator = SigVerificationDecorator{}

type SigVerificationDecorator struct {
	bank       types.BankKeeper
	underlying sdk.AnteDecorator
}

var _ sdk.AnteDecorator = SigVerificationDecorator{}

func NewSigVerificationDecorator(bk types.BankKeeper, underlying sdk.AnteDecorator) SigVerificationDecorator {
	if underlying == nil {
		panic("underlying ante decorator cannot be nil")
	}

	return SigVerificationDecorator{
		bank:       bk,
		underlying: underlying,
	}
}

func (d SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if msgs := tx.GetMsgs(); len(msgs) == 1 {
		msg, ok := msgs[0].(*types.MsgRegisterAccount)
		if !ok {
			return d.underlying.AnteHandle(ctx, tx, simulate, next)
		}

		address := types.GenerateAddress(msg.Channel, msg.Recipient, msg.Fallback)
		balance := d.bank.GetAllBalances(ctx, address)

		if balance.IsZero() || msg.Signer != address.String() {
			return d.underlying.AnteHandle(ctx, tx, simulate, next)
		}

		return next(ctx, tx, simulate)
	}

	return d.underlying.AnteHandle(ctx, tx, simulate, next)
}
