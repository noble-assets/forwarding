package v1

import (
	"strconv"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
)

// GetAllNumOfAccounts implements adapted legacy store logic from version 1.
// https://github.com/noble-assets/forwarding/blob/v1.x/x/forwarding/keeper/state.go#L25-L39
func GetAllNumOfAccounts(adapter storetypes.KVStore) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(adapter, types.NumOfAccountsPrefix)
	iterator := store.Iterator(nil, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}

// GetAllNumOfForwards implements adapted legacy store logic from version 1.
// https://github.com/noble-assets/forwarding/blob/v1.x/x/forwarding/keeper/state.go#L71-L85
func GetAllNumOfForwards(adapter storetypes.KVStore) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(adapter, types.NumOfForwardsPrefix)
	iterator := store.Iterator(nil, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}
