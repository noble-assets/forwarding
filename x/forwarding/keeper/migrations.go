package keeper

import (
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/noble-assets/forwarding/x/forwarding/migrations/v1"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	adapter := runtime.KVStoreAdapter(m.keeper.storeService.OpenKVStore(ctx))

	for channel, count := range v1.GetAllNumOfAccounts(adapter) {
		err := m.keeper.NumOfAccounts.Set(ctx, channel, count)
		if err != nil {
			return err
		}
	}

	for channel, count := range v1.GetAllNumOfForwards(adapter) {
		err := m.keeper.NumOfForwards.Set(ctx, channel, count)
		if err != nil {
			return err
		}
	}

	return nil
}
