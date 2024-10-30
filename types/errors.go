package types

import "cosmossdk.io/errors"

var (
	ErrInvalidAuthority = errors.Register(ModuleName, 1, "signer is not authority")
	ErrInvalidDenoms    = errors.Register(ModuleName, 2, "invalid allowed denoms")
)
