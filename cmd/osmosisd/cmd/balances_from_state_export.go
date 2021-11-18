package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

type DeriveSnapshot struct {
	NumberAccounts   uint64                       `json:"num_accounts"`
	PoolTokenCounter map[string]*PoolTokenCounter `json:"pool_token_counter"`
	Accounts         map[string]*DerivedAccount   `json:"accounts"`
}

// DerivedAccount provide fields of snapshot per account
type DerivedAccount struct {
	Address  string    `json:"address"`
	Balances sdk.Coins `json:"balance"`
}

func newDerivedAccount(address string) *DerivedAccount {
	return &DerivedAccount{
		Address:  address,
		Balances: sdk.Coins{},
	}
}

type PoolTokenCounter struct {
	PoolTotal sdk.Int `json:"pool_total"`
	CoinTotal sdk.Int `json:"coin_total"`
}

// ExportAirdropSnapshotCmd generates a snapshot.json from a provided exported genesis.json
func ExportDeriveBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-derive-balances [input-genesis-file] [output-snapshot-json]",
		Short: "Export a derive balances from a provided genesis export",
		Long: `Export a derive balances from a provided genesis export
Example:
	osmosisd export-derive-balances ../genesis.json ../snapshot.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPools := []string{
				"gamm/pool/1",
				"gamm/pool/3",
				"gamm/pool/6",
				"gamm/pool/498",
				"gamm/pool/13",
				"gamm/pool/5",
				"gamm/pool/4",
				"gamm/pool/497",
				"gamm/pool/15",
				"gamm/pool/10",
				"gamm/pool/9",
				"gamm/pool/22",
				"gamm/pool/42",
				"gamm/pool/8",
				"gamm/pool/7",
				"gamm/pool/2",
				"gamm/pool/561",
			}
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genesisFile := args[0]
			snapshotOutput := args[1]

			// Read genesis file
			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var doc tmtypes.GenesisDoc
			if err = tmjson.Unmarshal(byteValue, &doc); err != nil {
				return err
			}

			genState := make(map[string]json.RawMessage)
			if err = json.Unmarshal(doc.AppState, &genState); err != nil {
				panic(err)
			}

			snapshotAccs := map[string]*DerivedAccount{}
			pools := map[string]*PoolTokenCounter{}
			gammGenesis := gammtypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["gamm"], &gammGenesis)
			for _, any := range gammGenesis.Pools {
				var pool gammtypes.PoolI
				err := clientCtx.InterfaceRegistry.UnpackAny(any, &pool)
				if err != nil {
					panic(err)
				}
				tot := pool.GetTotalShares()
				for _, d := range targetPools {
					if d == tot.Denom {
						pools[tot.Denom] = &PoolTokenCounter{
							PoolTotal: tot.Amount,
							CoinTotal: sdk.ZeroInt(),
						}
					}
				}
			}

			bankGenesis := banktypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["bank"], &bankGenesis)
			for _, balance := range bankGenesis.Balances {
				address := balance.Address
				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				for _, c := range balance.Coins {
					for _, d := range targetPools {
						if c.Denom == d {
							acc.Balances = acc.Balances.Add(c)
							ptc := pools[c.Denom]
							ptc.CoinTotal = ptc.CoinTotal.Add(c.Amount)
							pools[c.Denom] = ptc
						}
					}
				}
				if acc.Balances.IsAllPositive() {
					snapshotAccs[address] = acc
				}
			}

			lockupGenesis := lockuptypes.GenesisState{}
			clientCtx.JSONMarshaler.MustUnmarshalJSON(genState["lockup"], &lockupGenesis)
			for _, lock := range lockupGenesis.Locks {
				address := lock.Owner

				acc, ok := snapshotAccs[address]
				if !ok {
					acc = newDerivedAccount(address)
				}

				for _, c := range lock.Coins {
					for _, d := range targetPools {
						if c.Denom == d {
							acc.Balances = acc.Balances.Add(c)
						}
					}
				}
				snapshotAccs[address] = acc
			}

			snapshot := DeriveSnapshot{
				NumberAccounts:   uint64(len(snapshotAccs)),
				PoolTokenCounter: pools,
				Accounts:         snapshotAccs,
			}

			// fmt.Printf("# accounts: %d\n", len(snapshotAccs))

			// export snapshot json
			snapshotJSON, err := json.MarshalIndent(snapshot, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}

			return ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
		},
	}

	// flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
