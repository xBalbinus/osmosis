package rate_limit

import (
	"golang.org/x/time/rate"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

/*
 If tx is greater than the max value of mempool, the user cannot put tx into the node.
 Since tx is propagated by p2p, it is difficult to block it in the load balancer.
 To alleviate the problem, add a rate limit to the logic.
 Only one tx per 2 seconds is allowed up to 5000 accounts that have recently requested tx.
*/
type RateLimitMiddleware struct {
}

var _ sdk.AnteDecorator = RateLimitMiddleware{}

var limiterMap = NewSizedMap(5, 1000, func() interface{} {
	// one per 2 seconds.
	return rate.NewLimiter(0.5, 1)
})

func NewRateLimitDecorator() RateLimitMiddleware {
	return RateLimitMiddleware{}
}

func (r RateLimitMiddleware) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}
	if simulate {
		return next(ctx, tx, simulate)
	}

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	for _, addr := range sigTx.GetSigners() {
		limiter, ok := limiterMap.Get(addr.String()).(*rate.Limiter)
		if !ok {
			return next(ctx, tx, simulate)
		}
		if !limiter.Allow() {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "rate limit has been reached. please wait and try again.")
		}
	}

	return next(ctx, tx, simulate)
}
