package rate_limit_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	ratelimit "github.com/osmosis-labs/osmosis/x/rate-limit"
)

func (suite *KeeperTestSuite) TestRateLimitDecorator() {
	suite.SetupTest(false)

	suite.ctx = suite.ctx.WithIsCheckTx(true)

	tx := legacytx.NewStdTx([]sdk.Msg{
		banktypes.NewMsgSend(acc1, acc2, sdk.NewCoins()),
	}, legacytx.NewStdFee(
		0, sdk.NewCoins(),
	), []legacytx.StdSignature{}, "")

	deco := ratelimit.NewRateLimitDecorator()
	antehandler := sdk.ChainAnteDecorators(deco)
	_, err := antehandler(suite.ctx, tx, false)
	suite.NoError(err)
	_, err = antehandler(suite.ctx, legacytx.NewStdTx([]sdk.Msg{
		banktypes.NewMsgSend(acc2, acc1, sdk.NewCoins()),
	}, legacytx.NewStdFee(
		0, sdk.NewCoins(),
	), []legacytx.StdSignature{}, ""), false)
	suite.NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second))
	_, err = antehandler(suite.ctx, tx, false)
	suite.Error(err)
	_, err = antehandler(suite.ctx.WithIsCheckTx(false), tx, false)
	suite.NoError(err)
	_, err = antehandler(suite.ctx.WithIsReCheckTx(true), tx, false)
	suite.NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Millisecond * 500))
	_, err = antehandler(suite.ctx, tx, false)
	suite.Error(err)
	_, err = antehandler(suite.ctx.WithIsCheckTx(false), tx, false)
	suite.NoError(err)
	_, err = antehandler(suite.ctx.WithIsReCheckTx(true), tx, false)
	suite.NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Millisecond * 500))
	_, err = antehandler(suite.ctx, tx, false)
	suite.NoError(err)
}
