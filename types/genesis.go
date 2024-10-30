package types

import (
	"errors"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		AllowedDenoms: []string{"*"},
	}
}

func (gen *GenesisState) Validate() error {
	if err := ValidateAllowedDenoms(gen.AllowedDenoms); err != nil {
		return err
	}

	for channel := range gen.NumOfAccounts {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}
	}

	for channel := range gen.NumOfForwards {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}
	}

	for channel, total := range gen.TotalForwarded {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}

		if _, err := sdk.ParseCoinsNormalized(total); err != nil {
			return errors.New("invalid coins")
		}
	}

	return nil
}

// ValidateAllowedDenoms checks if a specified denom list is valid.
// It ensures that if a wildcard "*" is present, it must be the only item.
// It also ensures non-empty entries.
func ValidateAllowedDenoms(denoms []string) error {
	if slices.Contains(denoms, "*") && len(denoms) > 1 {
		return errors.New("wildcard can only be present by itself")
	}

	for _, denom := range denoms {
		if strings.TrimSpace(denom) == "" {
			return errors.New("cannot allow empty denom")
		}
	}

	return nil
}
