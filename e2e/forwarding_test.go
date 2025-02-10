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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/jsonpb"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/icza/dyno"
	forwardingtypes "github.com/noble-assets/forwarding/v2/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/rly"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestRegisterOnNoble(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, sender, _, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.True(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err := noble.BankQueryAllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(1_000_000))), stats.TotalForwarded)
}

func TestRegisterOnNobleSignerlessly(t *testing.T) {
	t.Parallel()

	ctx, noble, _, _, _, sender, _, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	// NOTE: The keyName argument is intentionally left blank here. If
	//  everything is working correctly, this shouldn't error as we don't need
	//  to interact with the keyring.
	_, err := validator.ExecTx(ctx, "", "forwarding", "register-account-signerlessly", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.True(t, exists)
}

func TestRegisterViaTransfer(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, _, _, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	tx, err := gaia.SendIBCTransfer(ctx, "channel-0", receiver.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uatom",
		Amount:  math.NewInt(100_000),
	}, ibc.TransferOptions{
		Memo: fmt.Sprintf("{\"noble\":{\"forwarding\":{\"recipient\":\"%s\"}}}", receiver.FormattedAddress()),
	})
	require.NoError(t, err)
	fee := TxFee(t, ctx, gaia.Validators[0], tx.TxHash).AmountOf("uatom")

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.True(t, exists)

	balance, err := noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), "uatom")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000).Sub(fee), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uatom",
	}.IBCDenom(), math.NewInt(100_000))), stats.TotalForwarded)
}

func TestRegisterViaPacket(t *testing.T) {
	t.Skip()
}

func TestFrontRunAccount(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, sender, _, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.True(t, exists)

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err := noble.BankQueryAllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(1_000_000))), stats.TotalForwarded)
}

func TestClearAccount(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, rly, execReporter, sender, _, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	require.NoError(t, rly.StopRelayer(ctx, execReporter))

	address, exists := ForwardingAccount(t, ctx, validator, receiver, "")
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver, "")
	require.True(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	time.Sleep(10 * time.Minute)

	require.NoError(t, rly.StartRelayer(ctx, execReporter))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err := noble.GetBalance(ctx, sender.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := noble.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), balance)

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.True(t, receiverBalance.IsZero())

	_, err = validator.ExecTx(ctx, sender.KeyName(), "forwarding", "clear-account", address)
	require.NoError(t, err)
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err = noble.GetBalance(ctx, sender.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err = noble.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err = gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(2), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(2_000_000))), stats.TotalForwarded)
}

func TestFallbackAccount(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, rly, execReporter, sender, fallback, receiver := ForwardingSuite(t, nil)
	validator := noble.Validators[0]

	require.NoError(t, rly.StopRelayer(ctx, execReporter))

	address, exists := ForwardingAccount(t, ctx, validator, receiver, fallback.FormattedAddress())
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress(), fallback.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver, fallback.FormattedAddress())
	require.True(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	time.Sleep(10 * time.Minute)

	require.NoError(t, rly.StartRelayer(ctx, execReporter))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err := noble.GetBalance(ctx, sender.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	fallbackBalance, err := noble.GetBalance(ctx, fallback.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, fallbackBalance.IsZero())

	balance, err := noble.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), balance)

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.True(t, receiverBalance.IsZero())

	_, err = validator.ExecTx(ctx, sender.KeyName(), "forwarding", "clear-account", address, "--fallback")
	require.NoError(t, err)
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	senderBalance, err = noble.GetBalance(ctx, sender.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	fallbackBalance, err = noble.GetBalance(ctx, fallback.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1_000_000), fallbackBalance)

	balance, err = noble.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err = gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.True(t, receiverBalance.IsZero())

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(1_000_000))), stats.TotalForwarded)
}

func TestAllowedDenoms(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, sender, fallback, receiver := ForwardingSuite(t, &[]string{"uusdc"})
	validator := noble.Validators[0]

	res := ForwardingDenoms(t, ctx, validator)
	require.Len(t, res.AllowedDenoms, 1)
	require.Contains(t, res.AllowedDenoms, "uusdc")

	address, exists := ForwardingAccount(t, ctx, validator, receiver, fallback.FormattedAddress())
	require.False(t, exists)

	_, err := gaia.SendIBCTransfer(ctx, "channel-0", receiver.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uatom",
		Amount:  math.NewInt(100_000),
	}, ibc.TransferOptions{
		Memo: fmt.Sprintf("{\"noble\":{\"forwarding\":{\"recipient\":\"%s\",\"fallback\":\"%s\"}}}", receiver.FormattedAddress(), fallback.FormattedAddress()),
	})
	require.NoError(t, err)

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	_, exists = ForwardingAccount(t, ctx, validator, receiver, fallback.FormattedAddress())
	require.True(t, exists)

	balance, err := noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	uatom := transfertypes.DenomTrace{Path: "transfer/channel-0", BaseDenom: "uatom"}.IBCDenom()
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(uatom, math.NewInt(100_000))), balance)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	balance, err = noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(uatom, math.NewInt(100_000))), balance)

	_, err = validator.ExecTx(ctx, sender.KeyName(), "forwarding", "clear-account", address, "--fallback")
	require.NoError(t, err)
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, noble, gaia))

	balance, err = noble.BankQueryAllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	fallbackBalance, err := noble.BankQueryAllBalances(ctx, fallback.FormattedAddress())
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(uatom, math.NewInt(100_000))), fallbackBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", math.NewInt(1_000_000))), stats.TotalForwarded)
}

//

func ForwardingDenoms(t *testing.T, ctx context.Context, validator *cosmos.ChainNode) forwardingtypes.QueryDenomsResponse {
	raw, _, err := validator.ExecQuery(ctx, "forwarding", "denoms")
	require.NoError(t, err)

	var res forwardingtypes.QueryDenomsResponse
	require.NoError(t, json.Unmarshal(raw, &res))

	return res
}

func ForwardingAccount(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, receiver ibc.Wallet, fallback string) (address string, exists bool) {
	var raw []byte
	var err error

	if fallback == "" {
		raw, _, err = validator.ExecQuery(ctx, "forwarding", "address", "channel-0", receiver.FormattedAddress())
	} else {
		raw, _, err = validator.ExecQuery(ctx, "forwarding", "address", "channel-0", receiver.FormattedAddress(), fallback)
	}

	require.NoError(t, err)

	var res forwardingtypes.QueryAddressResponse
	require.NoError(t, json.Unmarshal(raw, &res))

	return res.Address, res.Exists
}

func ForwardingStats(t *testing.T, ctx context.Context, validator *cosmos.ChainNode) forwardingtypes.QueryStatsByChannelResponse {
	raw, _, err := validator.ExecQuery(ctx, "forwarding", "stats", "channel-0")
	require.NoError(t, err)

	var res forwardingtypes.QueryStatsByChannelResponse
	require.NoError(t, jsonpb.UnmarshalString(string(raw), &res))

	return res
}

type Fee struct {
	Amount sdk.Coins `json:"amount"`
}
type AuthInfo struct {
	Fee Fee `json:"fee"`
}
type Tx struct {
	AuthInfo AuthInfo `json:"auth_info"`
}
type TxResponse struct {
	Tx Tx `json:"tx"`
}

func TxFee(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, hash string) sdk.Coins {
	raw, _, err := validator.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err)

	var res TxResponse
	require.NoError(t, json.Unmarshal(raw, &res))

	return res.Tx.AuthInfo.Fee.Amount
}

func ForwardingSuite(t *testing.T, denoms *[]string) (ctx context.Context, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, relayer *rly.CosmosRelayer, execReporter *testreporter.RelayerExecReporter, sender ibc.Wallet, fallback ibc.Wallet, receiver ibc.Wallet) {
	ctx = context.Background()
	logger := zaptest.NewLogger(t)
	reporter := testreporter.NewNopReporter()
	execReporter = reporter.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	numValidators, numFullNodes := 1, 0

	factory := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		{
			Name:          "forwarding",
			Version:       "local",
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "forwarding",
				ChainID: "forwarding-1",
				Images: []ibc.DockerImage{
					{
						Repository: "noble-forwarding-simd",
						Version:    "local",
						UidGid:     "1025:1025",
					},
				},
				Bin:            "simd",
				Bech32Prefix:   "noble",
				Denom:          "uusdc",
				GasPrices:      "0uusdc",
				GasAdjustment:  5,
				TrustingPeriod: "504h",
				NoHostMount:    false,
				ModifyGenesis: func(cfg ibc.ChainConfig, bz []byte) ([]byte, error) {
					if denoms == nil {
						return bz, nil
					}

					gen := make(map[string]interface{})
					if err := json.Unmarshal(bz, &gen); err != nil {
						return nil, fmt.Errorf("failed to unmarshal genesis: %w", err)
					}

					if err := dyno.Set(gen, denoms, "app_state", "forwarding", "allowed_denoms"); err != nil {
						return nil, fmt.Errorf("failed to set forwarding allowed denoms in genesis: %w", err)
					}

					bz, err := json.Marshal(&gen)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal genesis: %w", err)
					}

					return bz, nil
				},
			},
		},
		{
			Name:          "gaia",
			Version:       "v14.2.0", // TODO(@john): Doesn't work on the v15 release line.
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
			ChainConfig: ibc.ChainConfig{
				ChainID: "cosmoshub-4",
			},
		},
	})

	chains, err := factory.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)
	gaia = chains[1].(*cosmos.CosmosChain)

	relayer = interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		logger,
	).Build(t, client, network).(*rly.CosmosRelayer)

	interchain := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(relayer, "rly").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: relayer,
			Path:    "transfer",
		})

	require.NoError(t, interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
	}))

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	require.NoError(t, relayer.StartRelayer(ctx, execReporter))

	wallets := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.NewInt(1_000_000), noble, noble, noble, gaia)
	sender, fallback, throwaway, receiver := wallets[0], wallets[1], wallets[2], wallets[3]

	require.NoError(t, noble.SendFunds(ctx, fallback.KeyName(), ibc.WalletAmount{
		Address: throwaway.FormattedAddress(),
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	return
}
