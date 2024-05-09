package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/jsonpb"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	forwardingtypes "github.com/noble-assets/forwarding/v2/x/forwarding/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/rly"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestForwarding_RegisterOnNoble(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, sender, receiver := ForwardingSuite(t)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
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

func TestForwarding_RegisterViaTransfer(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, _, receiver := ForwardingSuite(t)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
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

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
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

func TestForwarding_RegisterViaPacket(t *testing.T) {
	t.Skip()
}

func TestForwarding_FrontRunAccount(t *testing.T) {
	t.Parallel()

	ctx, noble, gaia, _, _, sender, receiver := ForwardingSuite(t)
	validator := noble.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	require.NoError(t, validator.BankSend(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  math.NewInt(1_000_000),
	}))

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
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

//

func ForwardingAccount(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, receiver ibc.Wallet) (address string, exists bool) {
	raw, _, err := validator.ExecQuery(ctx, "forwarding", "address", "channel-0", receiver.FormattedAddress())
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

func ForwardingSuite(t *testing.T) (ctx context.Context, noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, relayer *rly.CosmosRelayer, execReporter *testreporter.RelayerExecReporter, sender ibc.Wallet, receiver ibc.Wallet) {
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

	wallets := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.NewInt(1_000_000), noble, gaia)
	sender, receiver = wallets[0], wallets[1]

	return
}
