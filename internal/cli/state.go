package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var blockHeight int64
var accountAddress string

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Visualize the state trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		fmt.Printf("Fetching proof for address %s at block %d...\n", accountAddress, blockHeight)

		proofResult, err := client.GetProof(context.Background(), accountAddress, blockHeight)
		if err != nil {
			log.Fatalf("Failed to get account proof: %v", err)
		}

		fmt.Println("\n--- Raw RLP Encoded Trie Nodes ---")
		for i, rlpNode := range proofResult.AccountProof {
			fmt.Printf("Node %d: %x\n", i, rlpNode)
		}
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.Flags().Int64Var(&blockHeight, "block-height", 0, "Block height (required)")
	stateCmd.Flags().StringVar(&accountAddress, "account-address", "", "Account address to inspect (required)")
	stateCmd.MarkFlagRequired("block-height")
	stateCmd.MarkFlagRequired("account-address")
}
