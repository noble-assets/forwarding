package keeper

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
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

		return &types.MsgRegisterAccountResponse{Address: address.String()}, k.eventService.EventManager(ctx).Emit(ctx, &types.AccountRegistered{
			Address:   address.String(),
			Channel:   msg.Channel,
			Recipient: msg.Recipient,
			Fallback:  msg.Fallback,
		})
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

	return &types.MsgRegisterAccountResponse{Address: address.String()}, k.eventService.EventManager(ctx).Emit(ctx, &types.AccountRegistered{
		Address:   address.String(),
		Channel:   account.Channel,
		Recipient: account.Recipient,
		Fallback:  account.Fallback,
	})
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

	fallback, err := k.accountKeeper.AddressCodec().StringToBytes(account.Fallback)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to decode fallback address")
	}
	err = k.bankKeeper.SendCoins(ctx, address, fallback, balance)
	if err != nil {
		return nil, errors.New("failed to clear balance to fallback account")
	}

	return &types.MsgClearAccountResponse{}, k.eventService.EventManager(ctx).Emit(ctx, &types.AccountCleared{
		Address:   msg.Address,
		Recipient: account.Fallback,
	})
}

func (k *Keeper) SetAllowedDenoms(ctx context.Context, msg *types.MsgSetAllowedDenoms) (*types.MsgSetAllowedDenomsResponse, error) {
	if msg.Signer != k.authority {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAuthority, "expected %s, got %s", k.authority, msg.Signer)
	}

	if err := types.ValidateAllowedDenoms(msg.Denoms); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidDenoms, err.Error())
	}

	previousDenoms := k.GetAllowedDenoms(ctx)
	if err := k.AllowedDenoms.Clear(ctx, nil); err != nil {
		return nil, errors.New("failed to clear allowed denoms from state")
	}
	for _, denom := range msg.Denoms {
		err := k.AllowedDenoms.Set(ctx, denom)
		if err != nil {
			return nil, fmt.Errorf("failed to set %s as allowed denom in state", denom)
		}
	}

	return &types.MsgSetAllowedDenomsResponse{}, k.eventService.EventManager(ctx).Emit(ctx, &types.AllowedDenomsConfigured{
		PreviousDenoms: previousDenoms,
		CurrentDenoms:  msg.Denoms,
	})
}
