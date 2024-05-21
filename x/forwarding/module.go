package forwarding

import (
	"context"
	"encoding/json"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	modulev1 "github.com/noble-assets/forwarding/v2/api/noble/forwarding/module/v1"
	forwardingv1 "github.com/noble-assets/forwarding/v2/api/noble/forwarding/v1"
	"github.com/noble-assets/forwarding/v2/x/forwarding/client/cli"
	"github.com/noble-assets/forwarding/v2/x/forwarding/keeper"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
	"github.com/spf13/cobra"
)

// ConsensusVersion defines the current x/forwarding module consensus version.
const ConsensusVersion = 2

var (
	_ module.AppModuleBasic      = AppModule{}
	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ appmodule.HasEndBlocker    = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
)

//

type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, cfg client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesis.Validate()
}

//

type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
	}
}

func (AppModule) IsOnePerModuleType() {}

func (AppModule) IsAppModule() {}

func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (m AppModule) EndBlock(ctx context.Context) error {
	m.keeper.ExecuteForwards(ctx)
	return nil
}

func (m AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesis)

	InitGenesis(ctx, m.keeper, genesis)
}

func (m AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := ExportGenesis(ctx, m.keeper)
	return cdc.MustMarshalJSON(genesis)
}

func (m AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), m.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), m.keeper)

	migrator := keeper.NewMigrator(m.keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, migrator.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/forwarding from version 1 to 2: %v", err))
	}
}

//

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: forwardingv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterAccount",
					Use:       "register-account [channel] [recipient] (fallback)",
					Short:     "Register a forwarding account for a channel and recipient",
					Long:      "Register a forwarding account for a channel and recipient, with an optional fallback address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "channel"},
						{ProtoField: "recipient"},
						{ProtoField: "fallback", Optional: true},
					},
				},
				{
					RpcMethod: "ClearAccount",
					Use:       "clear-account [address] (--fallback)",
					Short:     "Manually clear funds inside forwarding account",
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"fallback": {
							Usage: "Clear funds to fallback address, if exists",
						},
					},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod: "SetAllowedDenoms",
					Use:       "set-allowed-denoms [denoms ...]",
					Short:     "Set the list of denoms that are allowed to be forwarded",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{
						ProtoField: "denoms",
						Varargs:    true,
					}},
				},
			},
		},
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: forwardingv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Denoms",
					Use:       "denoms",
					Short:     "Query denoms that are allowed to be forwarded",
				},
				{
					RpcMethod: "Address",
					Use:       "address [channel] [recipient] (fallback)",
					Short:     "Query forwarding address by channel and recipient",
					Long:      "Query forwarding address by channel and recipient, with an optional fallback address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "channel"},
						{ProtoField: "recipient"},
						{ProtoField: "fallback", Optional: true},
					},
				},
				// NOTE: We combine the Stats and StatsByChannel methods together in a custom command.
				{
					RpcMethod: "Stats",
					Skip:      true,
				},
				{
					RpcMethod: "StatsByChannel",
					Skip:      true,
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}

func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

//

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config           *modulev1.Module
	Cdc              codec.Codec
	Logger           log.Logger
	StoreService     store.KVStoreService
	TransientService store.TransientStoreService
	HeaderService    header.Service
	EventService     event.Service

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper      *keeper.Keeper
	Module      appmodule.AppModule
	Restriction banktypes.SendRestrictionFn
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	if in.Config.Authority == "" {
		panic("authority for x/forwarding module must be set")
	}

	authority := authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	k := keeper.NewKeeper(
		in.Cdc,
		in.Logger,
		in.StoreService,
		in.TransientService,
		in.HeaderService,
		in.EventService,
		authority.String(),
		in.AccountKeeper,
		in.BankKeeper,
		nil,
		nil,
	)
	m := NewAppModule(k)

	return ModuleOutputs{Keeper: k, Module: m, Restriction: k.SendRestrictionFn}
}
