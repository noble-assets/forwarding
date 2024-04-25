package forwarding

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
<<<<<<< HEAD
	"github.com/noble-assets/forwarding/x/forwarding/client/cli"
	"github.com/noble-assets/forwarding/x/forwarding/keeper"
	"github.com/noble-assets/forwarding/x/forwarding/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
=======
	modulev1 "github.com/noble-assets/forwarding/v2/api/noble/forwarding/module/v1"
	forwardingv1 "github.com/noble-assets/forwarding/v2/api/noble/forwarding/v1"
	"github.com/noble-assets/forwarding/v2/x/forwarding/client/cli"
	"github.com/noble-assets/forwarding/v2/x/forwarding/keeper"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
	"github.com/spf13/cobra"
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
)

var (
	_ module.EndBlockAppModule = AppModule{}
	_ module.AppModuleBasic    = AppModuleBasic{}
)

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

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesis)

	InitGenesis(ctx, am.keeper, genesis)

	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genesis)
}

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (AppModule) Route() sdk.Route { return sdk.Route{} }

func (AppModule) QuerierRoute() string { return types.ModuleName }

func (AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier { return nil }

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	am.keeper.ExecuteForwards(ctx)

	return []abci.ValidatorUpdate{}
}

//

type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
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

func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return cli.GetTxCmd() }

<<<<<<< HEAD
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return cli.GetQueryCmd() }
=======
func (AppModule) IsOnePerModuleType() {}

func (AppModule) IsAppModule() {}

func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (m AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	m.keeper.ExecuteForwards(ctx)

	return nil, nil
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
}

//

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: forwardingv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "RegisterAccount",
					Use:            "register-account [channel] [recipient]",
					Short:          "Register a forwarding account for a channel and recipient",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "channel"}, {ProtoField: "recipient"}},
				},
				{
					RpcMethod:      "ClearAccount",
					Use:            "clear-account [address]",
					Short:          "Manually clear funds inside forwarding account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
			},
		},
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: forwardingv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Address",
					Use:            "address [channel] [recipient]",
					Short:          "Query forwarding address by channel and recipient",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "channel"}, {ProtoField: "recipient"}},
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
	StoreService     store.KVStoreService
	TransientService store.TransientStoreService
	Logger           log.Logger

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(
		in.Cdc,
		in.Logger,
		in.StoreService,
		in.TransientService,
		in.AccountKeeper,
		in.BankKeeper,
		nil,
		nil,
	)
	m := NewAppModule(k)

	return ModuleOutputs{Keeper: k, Module: m}
}
>>>>>>> 8ab8bfa (feat: add general stats query (#5))
