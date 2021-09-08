package cmd

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagTestHardForkChainUpdateHeight = "test-hardfork-chain-update-height"
	FlagTestHardForkValPubKeys        = "test-hardfork-val-pub-keys"
)

func FlagSetTestHardForkUpdate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Int64(FlagTestHardForkChainUpdateHeight, 0, "")
	fs.StringSlice(FlagTestHardForkValPubKeys, []string{}, "")
	return fs
}
