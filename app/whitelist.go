package app

import (
	"encoding/csv"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

// Every asset with a liquid osmo pairing pool on Osmosis, as of 12/01/21
// Notably, Tick is not on this list because the osmo pool has $76 of liquidity.
// Cheq'd and KRT are also not on this, due to neither having osmo pairings.
var asset_data = `
atom,ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2,1
`

func whitelistInitial(ctx sdk.Context, app *OsmosisApp) []types.FeeToken {
	r := csv.NewReader(strings.NewReader(asset_data))
	assets, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	feeTokens := make([]types.FeeToken, 0, len(assets))
	for _, asset := range assets {
		base10 := 10
		bitLen := 64
		poolIdStr := strings.TrimSpace(asset[2])
		poolId, err := strconv.ParseUint(poolIdStr, base10, bitLen)
		if err != nil {
			panic(err)
		}

		feeToken := types.FeeToken{
			Denom:  asset[1],
			PoolID: poolId,
		}

		feeTokens = append(feeTokens, feeToken)
	}
	return feeTokens
}
