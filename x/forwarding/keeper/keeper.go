package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
)

type Keeper struct {
	cdc              codec.Codec
	logger           log.Logger
	storeService     store.KVStoreService
	transientService store.TransientStoreService
	headerService    header.Service
	eventService     event.Service

	authority string

	Schema         collections.Schema
	AllowedDenoms  collections.KeySet[string]
	NumOfAccounts  collections.Map[string, uint64]
	NumOfForwards  collections.Map[string, uint64]
	TotalForwarded collections.Map[string, string]

	TransientSchema collections.Schema
	PendingForwards collections.Map[string, types.ForwardingAccount]

	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	channelKeeper  types.ChannelKeeper
	transferKeeper types.TransferKeeper
}

func NewKeeper(
	cdc codec.Codec,
	logger log.Logger,
	storeService store.KVStoreService,
	transientService store.TransientStoreService,
	headerService header.Service,
	eventService event.Service,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	transferKeeper types.TransferKeeper,
) *Keeper {
	builder := collections.NewSchemaBuilder(storeService)
	transientBuilder := collections.NewSchemaBuilderFromAccessor(transientService.OpenTransientStore)

	keeper := &Keeper{
		cdc:              cdc,
		logger:           logger,
		storeService:     storeService,
		transientService: transientService,
		headerService:    headerService,
		eventService:     eventService,

		authority: authority,

		AllowedDenoms:  collections.NewKeySet(builder, types.AllowedDenomsPrefix, "allowed_denoms", collections.StringKey),
		NumOfAccounts:  collections.NewMap(builder, types.NumOfAccountsPrefix, "num_of_accounts", collections.StringKey, collections.Uint64Value),
		NumOfForwards:  collections.NewMap(builder, types.NumOfForwardsPrefix, "num_of_forwards", collections.StringKey, collections.Uint64Value),
		TotalForwarded: collections.NewMap(builder, types.TotalForwardedPrefix, "total_forwarded", collections.StringKey, collections.StringValue),

		PendingForwards: collections.NewMap(transientBuilder, types.PendingForwardsPrefix, "pending_forwards", collections.StringKey, codec.CollValue[types.ForwardingAccount](cdc)),

		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		channelKeeper:  channelKeeper,
		transferKeeper: transferKeeper,
	}

	schema, err := builder.Build()
	if err != nil {
		panic(err)
	}
	transientSchema, err := transientBuilder.Build()
	if err != nil {
		panic(err)
	}

	keeper.Schema = schema
	keeper.TransientSchema = transientSchema
	return keeper
}

// IsAllowedDenom checks if a specific denom is allowed to be forwarded.
func (k *Keeper) IsAllowedDenom(ctx context.Context, denom string) bool {
	has, _ := k.AllowedDenoms.Has(ctx, "*")
	if has {
		return true
	}

	has, _ = k.AllowedDenoms.Has(ctx, denom)
	return has
}

// ExecuteForwards is an end block hook that clears all pending forwards from transient state.
func (k *Keeper) ExecuteForwards(ctx context.Context) {
	forwards := k.GetPendingForwards(ctx)
	if len(forwards) > 0 {
		k.Logger().Info(fmt.Sprintf("executing %d automatic forward(s)", len(forwards)))
	}

	for _, forward := range forwards {
		channel, _ := k.channelKeeper.GetChannel(sdk.UnwrapSDKContext(ctx), transfertypes.PortID, forward.Channel)
		if channel.State != channeltypes.OPEN {
			k.Logger().Error("skipped automatic forward due to non open channel", "channel", forward.Channel, "address", forward.GetAddress().String(), "state", channel.State.String())
			continue
		}

		balances := k.bankKeeper.GetAllBalances(ctx, forward.GetAddress())

		for _, balance := range balances {
			if !k.IsAllowedDenom(ctx, balance.Denom) {
				continue
			}

			timeout := uint64(k.headerService.GetHeaderInfo(ctx).Time.UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp
			_, err := k.transferKeeper.Transfer(ctx, &transfertypes.MsgTransfer{
				SourcePort:       transfertypes.PortID,
				SourceChannel:    forward.Channel,
				Token:            balance,
				Sender:           forward.Address,
				Receiver:         forward.Recipient,
				TimeoutHeight:    clienttypes.ZeroHeight(),
				TimeoutTimestamp: timeout,
				Memo:             "",
			})
			if err != nil {
				// TODO: Consider moving to persistent store in order to retry in future blocks?
				k.Logger().Error("unable to execute automatic forward", "channel", forward.Channel, "address", forward.GetAddress().String(), "amount", balance.String(), "err", err)
			} else {
				k.IncrementNumOfForwards(ctx, forward.Channel)
				k.IncrementTotalForwarded(ctx, forward.Channel, balance)
			}
		}
	}

	// NOTE: As pending forwards are stored in transient state, they are automatically cleared at the end of the block lifecycle. No further action is required.
}

// SendRestrictionFn checks every transfer executed on the Noble chain to see if
// the recipient is a forwarding account, allowing us to mark accounts for clearing.
func (k *Keeper) SendRestrictionFn(ctx context.Context, fromAddr, toAddr sdk.AccAddress, _ sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	rawAccount := k.accountKeeper.GetAccount(ctx, toAddr)
	if rawAccount == nil {
		return toAddr, nil
	}

	account, ok := rawAccount.(*types.ForwardingAccount)
	if !ok {
		return toAddr, nil
	}

	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, account.Channel)
	if !fromAddr.Equals(escrow) {
		k.SetPendingForward(ctx, account)
	}

	return toAddr, nil
}

// SetIBCKeepers allows us to set the relevant IBC keepers post dependency
// injection, as IBC doesn't support dependency injection yet.
func (k *Keeper) SetIBCKeepers(channelKeeper types.ChannelKeeper, transferKeeper types.TransferKeeper) {
	k.channelKeeper = channelKeeper
	k.transferKeeper = transferKeeper
}

func (k *Keeper) Logger() log.Logger {
	return k.logger.With("module", types.ModuleName)
}
