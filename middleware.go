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
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/noble-assets/forwarding/v2/keeper"
	"github.com/noble-assets/forwarding/v2/types"
)

var _ porttypes.IBCModule = &Middleware{}

type Middleware struct {
	app porttypes.IBCModule

	authKeeper types.AccountKeeper
	keeper     *keeper.Keeper
}

func NewMiddleware(app porttypes.IBCModule, authKeeper types.AccountKeeper, keeper *keeper.Keeper) Middleware {
	return Middleware{app: app, authKeeper: authKeeper, keeper: keeper}
}

func (m Middleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return m.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

func (m Middleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return m.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

func (m Middleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return m.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

func (m Middleware) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return m.app.OnChanOpenConfirm(ctx, portID, channelID)
}

func (m Middleware) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return m.app.OnChanCloseInit(ctx, portID, channelID)
}

func (m Middleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return m.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the porttypes.IBCModule interface.
func (m Middleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	// This middleware is intended to be integrated with "transfer" ICS-20
	// channels. With this middleware, two packets exist can be sent on these
	// channels, namely "FungibleTokenPacketData" and "RegisterAccountData".
	//
	// When receiving a "FungibleTokenPacketData" packet, we first check the
	// memo field. If the memo field contains registration data, we first
	// register a new forwarding account before continuing. We additionally
	// need to check if the recipient of the token transfer is a forwarding
	// account, as we then mark it for forwarding at the end of the block
	// lifecycle.
	//
	// When receiving a "RegisterAccountData" packet, we simply register a new
	// forwarding account.

	var transferData transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &transferData); err == nil {
		var memo types.RegisterAccountMemo
		if err := types.ModuleCdc.UnmarshalJSON([]byte(transferData.GetMemo()), &memo); err == nil {
			if memo.Noble != nil && memo.Noble.Forwarding != nil {
				channel := packet.DestinationChannel
				if memo.Noble.Forwarding.Channel != "" {
					channel = memo.Noble.Forwarding.Channel
				}

				req := &types.MsgRegisterAccount{
					Signer:    authtypes.NewModuleAddress(types.ModuleName).String(),
					Recipient: memo.Noble.Forwarding.Recipient,
					Channel:   channel,
					Fallback:  memo.Noble.Forwarding.Fallback,
				}

				_, err := m.keeper.RegisterAccount(ctx, req)
				if err != nil {
					return channeltypes.NewErrorAcknowledgement(err)
				}
			}
		}

		receiver, err := m.authKeeper.AddressCodec().StringToBytes(transferData.Receiver)
		if err != nil {
			return m.app.OnRecvPacket(ctx, packet, relayer)
		}

		rawAccount := m.authKeeper.GetAccount(ctx, receiver)
		if rawAccount == nil {
			return m.app.OnRecvPacket(ctx, packet, relayer)
		}

		account, ok := rawAccount.(*types.ForwardingAccount)
		if ok {
			m.keeper.SetPendingForward(ctx, account)
		}

		return m.app.OnRecvPacket(ctx, packet, relayer)
	}

	var data types.RegisterAccountData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return m.app.OnRecvPacket(ctx, packet, relayer)
	}

	channel := packet.DestinationChannel
	if data.Channel != "" {
		channel = data.Channel
	}

	req := &types.MsgRegisterAccount{
		Recipient: data.Recipient,
		Channel:   channel,
		Fallback:  data.Fallback,
	}

	res, err := m.keeper.RegisterAccount(ctx, req)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	} else {
		return channeltypes.NewResultAcknowledgement([]byte(res.Address))
	}
}

func (m Middleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return m.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

func (m Middleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return m.app.OnTimeoutPacket(ctx, packet, relayer)
}
