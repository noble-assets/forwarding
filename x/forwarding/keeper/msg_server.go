package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) RegisterAccount(ctx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	if !channeltypes.IsValidChannelID(msg.Channel) {
		return nil, errors.New("invalid channel")
	}

	if msg.Fallback != "" {
		if _, err := k.accountKeeper.AddressCodec().StringToBytes(msg.Fallback); err != nil {
			return nil, errors.New("invalid fallback address")
		}
	}
	address := types.GenerateAddress(msg.Channel, msg.Recipient, msg.Fallback)

	channel, found := k.channelKeeper.GetChannel(sdk.UnwrapSDKContext(ctx), transfertypes.PortID, msg.Channel)
	if !found {
		return nil, fmt.Errorf("channel does not exist: %s", msg.Channel)
	}
	if channel.State != channeltypes.OPEN {
		return nil, fmt.Errorf("channel is not open: %s, %s", msg.Channel, channel.State)
	}

	if k.accountKeeper.HasAccount(ctx, address) {
		rawAccount := k.accountKeeper.GetAccount(ctx, address)
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
				Fallback:    msg.Fallback,
			}
			k.accountKeeper.SetAccount(ctx, rawAccount)

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

	base := k.accountKeeper.NewAccountWithAddress(ctx, address)
	account := types.ForwardingAccount{
		BaseAccount: authtypes.NewBaseAccount(base.GetAddress(), base.GetPubKey(), base.GetAccountNumber(), base.GetSequence()),
		Channel:     msg.Channel,
		Recipient:   msg.Recipient,
		CreatedAt:   k.headerService.GetHeaderInfo(ctx).Height,
		Fallback:    msg.Fallback,
	}

	k.accountKeeper.SetAccount(ctx, &account)
	k.IncrementNumOfAccounts(ctx, msg.Channel)

	return &types.MsgRegisterAccountResponse{Address: address.String()}, nil
}

func (k *Keeper) ClearAccount(ctx context.Context, msg *types.MsgClearAccount) (*types.MsgClearAccountResponse, error) {
	address, err := k.accountKeeper.AddressCodec().StringToBytes(msg.Address)
	if err != nil {
		return nil, errors.New("invalid account address")
	}

	rawAccount := k.accountKeeper.GetAccount(ctx, address)
	if rawAccount == nil {
		return nil, errors.New("account does not exist")
	}
	account, ok := rawAccount.(*types.ForwardingAccount)
	if !ok {
		return nil, errors.New("account is not a forwarding account")
	}

	balance := k.bankKeeper.GetAllBalances(ctx, address)
	if balance.IsZero() {
		return nil, errors.New("account does not require clearing")
	}

	if !msg.Fallback || account.Fallback == "" {
		k.SetPendingForward(ctx, account)
		return &types.MsgClearAccountResponse{}, nil
	}

	fallback, _ := k.accountKeeper.AddressCodec().StringToBytes(account.Fallback)
	err = k.bankKeeper.SendCoins(ctx, address, fallback, balance)
	if err != nil {
		return nil, errors.New("failed to clear balance to fallback account")
	}

	return &types.MsgClearAccountResponse{}, nil
}
