package cli

import (
	"log"

	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var blockHeight int64

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Visualize the state trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		block, err := client.GetBlockByNumber(cmd.Context(), blockHeight)
		if err != nil {
			log.Fatalf("Failed to get block: %v", err)
		}

		log.Printf("Block %d state root: %s", blockHeight, block.Root().Hex())
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.Flags().Int64Var(&blockHeight, "MarkFlagRequired", 0, "Block height to visualize the state trie (required)")
	stateCmd.MarkFlagRequired("block-height")
}
