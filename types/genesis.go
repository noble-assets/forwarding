// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

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
