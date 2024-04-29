package simapp

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	forwardingkeeper "github.com/noble-assets/forwarding/v2/x/forwarding/keeper"

	_ "cosmossdk.io/api/cosmos/tx/config/v1"               // import for side-effects
	_ "cosmossdk.io/x/upgrade"                             // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth"                // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"      // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"                // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution"        // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/params"              // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"             // import for side-effects
	_ "github.com/noble-assets/forwarding/v2/x/forwarding" // import for side-effects
)

var DefaultNodeHome string

//go:embed app.yaml
var AppConfigYAML []byte

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// Cosmos SDK Modules
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	DistributionKeeper    distributionkeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	// IBC Modules
	CapabilityKeeper     *capabilitykeeper.Keeper
	IBCKeeper            *ibckeeper.Keeper
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	TransferKeeper       transferkeeper.Keeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	// Custom Modules
	ForwardingKeeper *forwardingkeeper.Keeper
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".simapp")
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appconfig.LoadYAML(AppConfigYAML),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			},
		),
	)
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*SimApp, error) {
	var (
		app        = &SimApp{}
		appBuilder *runtime.AppBuilder
	)

	if err := depinject.Inject(
		depinject.Configs(
			AppConfig(),
			depinject.Supply(
				logger,
				appOpts,
			),
		),
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		// Cosmos SDK Modules
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.ConsensusParamsKeeper,
		&app.DistributionKeeper,
		&app.ParamsKeeper,
		&app.StakingKeeper,
		&app.UpgradeKeeper,
		// Custom Modules
		&app.ForwardingKeeper,
	); err != nil {
		return nil, err
	}

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	if err := app.RegisterIBCModules(); err != nil {
		panic(err)
	}

	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *SimApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *SimApp) SimulationManager() *module.SimulationManager {
	return nil
}

func (app *SimApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	key, _ := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	return key
}

func (app *SimApp) GetMemKey(memKey string) *storetypes.MemoryStoreKey {
	key, _ := app.UnsafeFindStoreKey(memKey).(*storetypes.MemoryStoreKey)
	return key
}

func (app *SimApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

func (app *SimApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}
