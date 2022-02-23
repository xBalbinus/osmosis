package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an instance of MsgServer
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) SuperfluidDelegate(goCtx context.Context, msg *types.MsgSuperfluidDelegate) (*types.MsgSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidDelegate(ctx, msg.Sender, msg.LockId, msg.ValAddr)
	return &types.MsgSuperfluidDelegateResponse{}, err
}

func (server msgServer) SuperfluidUndelegate(goCtx context.Context, msg *types.MsgSuperfluidUndelegate) (*types.MsgSuperfluidUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUndelegate(ctx, msg.Sender, msg.LockId)
	return &types.MsgSuperfluidUndelegateResponse{}, err
}

// func (server msgServer) SuperfluidRedelegate(goCtx context.Context, msg *types.MsgSuperfluidRedelegate) (*types.MsgSuperfluidRedelegateResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	err := server.keeper.SuperfluidRedelegate(ctx, msg.Sender, msg.LockId, msg.NewValAddr)
// 	return &types.MsgSuperfluidRedelegateResponse{}, err
// }

func (server msgServer) SuperfluidUnbondLock(goCtx context.Context, msg *types.MsgSuperfluidUnbondLock) (
	*types.MsgSuperfluidUnbondLockResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUnbondLock(ctx, msg.LockId, msg.Sender)
	return &types.MsgSuperfluidUnbondLockResponse{}, err
}

func (server msgServer) LockAndSuperfluidDelegate(goCtx context.Context, msg *types.MsgLockAndSuperfluidDelegate) (*types.MsgLockAndSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	lockup, err := server.keeper.lk.LockTokens(ctx, sender, msg.Coins, server.keeper.sk.GetParams(ctx).UnbondingTime)
	if err != nil {
		return &types.MsgLockAndSuperfluidDelegateResponse{}, err
	}

	superfluidDelegateMsg := types.MsgSuperfluidDelegate{
		Sender:  msg.Sender,
		LockId:  lockup.ID,
		ValAddr: msg.ValAddr,
	}

	_, err = server.SuperfluidDelegate(goCtx, &superfluidDelegateMsg)
	return &types.MsgLockAndSuperfluidDelegateResponse{
		ID: lockup.ID,
	}, err
}
