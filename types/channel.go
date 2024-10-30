package types

import (
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

func ParseChainId(rawClientState exported.ClientState) string {
	switch clientState := rawClientState.(type) {
	case *tendermint.ClientState:
		return clientState.ChainId
	default:
		return "UNKNOWN"
	}
}
