package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/noble-assets/forwarding/x/forwarding/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) RegisterAccount(ctx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	if !channeltypes.IsValidChannelID(msg.Channel) {
		return nil, errors.New("invalid channel")
	}

	address := types.GenerateAddress(msg.Channel, msg.Recipient)

	channel, found := k.channelKeeper.GetChannel(sdk.UnwrapSDKContext(ctx), transfertypes.PortID, msg.Channel)
	if !found {
		return nil, fmt.Errorf("channel does not exist: %s", msg.Channel)
	}
	if channel.State != channeltypes.OPEN {
		return nil, fmt.Errorf("channel is not open: %s, %s", msg.Channel, channel.State)
	}

	if k.authKeeper.HasAccount(ctx, address) {
		rawAccount := k.authKeeper.GetAccount(ctx, address)
		if rawAccount.GetPubKey() != nil || rawAccount.GetSequence() != 0 {
			return nil, fmt.Errorf("attempting to register an existing user account with address: %s", address.String())
		}

		switch account := rawAccount.(type) {
		case *authtypes.BaseAccount:
			rawAccount = &types.ForwardingAccount{
				BaseAccount: account,
				Channel:     msg.Channel,
				Recipient:   msg.Recipient,
				CreatedAt:   k.headerService.GetHeaderInfo(ctx).Height,
			}
			k.authKeeper.SetAccount(ctx, rawAccount)

			k.IncrementNumOfAccounts(ctx, msg.Channel)
		case *types.ForwardingAccount:
			return nil, errors.New("account has already been registered")
		default:
			return nil, fmt.Errorf("unsupported account type: %T", rawAccount)
		}

		if !k.bankKeeper.GetAllBalances(ctx, address).IsZero() {
			account, ok := rawAccount.(*types.ForwardingAccount)
			if ok {
				k.SetPendingForward(ctx, account)
			}
		}

		return &types.MsgRegisterAccountResponse{Address: address.String()}, nil
	}

	base := k.authKeeper.NewAccountWithAddress(ctx, address)
	account := types.ForwardingAccount{
		BaseAccount: authtypes.NewBaseAccount(base.GetAddress(), base.GetPubKey(), base.GetAccountNumber(), base.GetSequence()),
		Channel:     msg.Channel,
		Recipient:   msg.Recipient,
		CreatedAt:   k.headerService.GetHeaderInfo(ctx).Height,
	}

	k.authKeeper.SetAccount(ctx, &account)
	k.IncrementNumOfAccounts(ctx, msg.Channel)

	return &types.MsgRegisterAccountResponse{Address: address.String()}, nil
}

func (k *Keeper) ClearAccount(ctx context.Context, msg *types.MsgClearAccount) (*types.MsgClearAccountResponse, error) {
	address, err := k.authKeeper.AddressCodec().StringToBytes(msg.Address)
	if err != nil {
		return nil, errors.New("invalid account address")
	}

	rawAccount := k.authKeeper.GetAccount(ctx, address)
	if rawAccount == nil {
		return nil, errors.New("account does not exist")
	}
	account, ok := rawAccount.(*types.ForwardingAccount)
	if !ok {
		return nil, errors.New("account is not a forwarding account")
	}

	if k.bankKeeper.GetAllBalances(ctx, address).IsZero() {
		return nil, errors.New("account does not require clearing")
	}

	k.SetPendingForward(ctx, account)

	return &types.MsgClearAccountResponse{}, nil
}
