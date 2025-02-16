package simapp

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/noble-assets/forwarding/v2"
)

func (app *SimApp) RegisterLegacyModules() error {
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(capabilitytypes.StoreKey),
		storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey),
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(transfertypes.StoreKey),
	); err != nil {
		return err
	}

	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(clienttypes.ParamKeyTable().RegisterParamSet(&connectiontypes.Params{}))
	app.ParamsKeeper.Subspace(transfertypes.ModuleName).WithKeyTable(transfertypes.ParamKeyTable())

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		app.appCodec,
		app.GetKey(capabilitytypes.StoreKey),
		app.GetMemKey(capabilitytypes.MemStoreKey),
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		app.GetKey(ibcexported.StoreKey),
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		"noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
	)

	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(transfertypes.ModuleName)
	app.TransferKeeper = transferkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(transfertypes.StoreKey),
		app.GetSubspace(transfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		"noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
	)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = forwarding.NewMiddleware(transferStack, app.AccountKeeper, app.ForwardingKeeper)

	ibcRouter := porttypes.NewRouter().AddRoute(transfertypes.ModuleName, transferStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	app.ForwardingKeeper.SetIBCKeepers(app.IBCKeeper.ChannelKeeper, app.TransferKeeper)

	return app.RegisterModules(
		capability.NewAppModule(app.appCodec, *app.CapabilityKeeper, true),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		tendermint.NewAppModule(),
	)
}
