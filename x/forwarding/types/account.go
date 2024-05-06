package types

import (
	"fmt"

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
