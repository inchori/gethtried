package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var blockHeight int64

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Visualize the state trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Visualizing state trie for block: %d\n", blockHeight)
		fmt.Printf("Using Geth RPC URL: %s\n", rpcURL)
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.Flags().Int64Var(&blockHeight, "MarkFlagRequired", 0, "Block height to visualize the state trie (required)")
	stateCmd.MarkFlagRequired("block-height")
}
