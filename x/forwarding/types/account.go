package types

import (
	"bytes"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
)

var (
	_ authtypes.AccountI       = &ForwardingAccount{}
	_ authtypes.GenesisAccount = &ForwardingAccount{}
)

func GenerateAddress(channel string, recipient string) sdk.AccAddress {
	bz := []byte(channel + recipient)
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
