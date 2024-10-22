package forwarding

import (
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding/types"
)

// SigVerificationGasConsumer is a wrapper around the default provided by the
// Cosmos SDK that supports forwarding account public keys.
func SigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	switch sig.PubKey.(type) {
	case *types.ForwardingPubKey:
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}

//

type SigVerificationDecorator struct {
	underlying ante.SigVerificationDecorator
	bank       types.BankKeeper
}

var _ sdk.AnteDecorator = SigVerificationDecorator{}

func NewSigVerificationDecorator(ak ante.AccountKeeper, bk types.BankKeeper, signModeHandler *txsigning.HandlerMap) SigVerificationDecorator {
	return SigVerificationDecorator{
		underlying: ante.NewSigVerificationDecorator(ak, signModeHandler),
		bank:       bk,
	}
}

func (d SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if msgs := tx.GetMsgs(); len(msgs) == 1 {
		msg, ok := msgs[0].(*types.MsgRegisterAccount)
		if !ok {
			return d.underlying.AnteHandle(ctx, tx, simulate, next)
		}

		address := types.GenerateAddress(msg.Channel, msg.Recipient, msg.Fallback)
		balance := d.bank.GetAllBalances(ctx, address)

		if balance.IsZero() || msg.Signer != address.String() {
			return d.underlying.AnteHandle(ctx, tx, simulate, next)
		}

		return next(ctx, tx, simulate)
	}

	return d.underlying.AnteHandle(ctx, tx, simulate, next)
}
