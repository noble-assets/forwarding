package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterAccount{}, "noble/forwarding/RegisterAccount", nil)
	cdc.RegisterConcrete(&MsgClearAccount{}, "noble/forwarding/ClearAccount", nil)
	cdc.RegisterConcrete(&MsgSetAllowedDenoms{}, "noble/forwarding/SetAllowedDenoms", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.AccountI)(nil), &ForwardingAccount{})
	registry.RegisterImplementations((*authtypes.GenesisAccount)(nil), &ForwardingAccount{})

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterAccount{},
		&MsgClearAccount{},
		&MsgSetAllowedDenoms{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
