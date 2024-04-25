package types

import (
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
)

func ParseChainId(rawClientState exported.ClientState) string {
	switch clientState := rawClientState.(type) {
	case *tendermint.ClientState:
		return clientState.ChainId
	default:
		return "UNKNOWN"
	}
}
