package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

func MigratePoolStruct(ctx sdk.Context, k Keeper) {
	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	for _, legacyPool := range pools {
		// delete old pool
		k.DeletePool(ctx, legacyPool.GetId())

		balancerPool, ok := legacyPool.(*balancer.BalancerPool)
		if !ok {
			panic(fmt.Errorf(""))
		}
		k.SetPool(ctx, balancerPool)
	}
}
