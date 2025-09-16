package cli

import (
	"context"
	"log"

	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var storageSlot int64

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Visualize the storage trie for a specific account at a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		storageProof, err := client.GetStorageProof(context.Background(), accountAddress, storageSlot, blockHeight)
		if err != nil {
			log.Fatalf("Failed to get storage proof: %v", err)
		}

		log.Printf("Successfully got %d proof nodes for storage slot %d of account %s at block %d.\n", len(storageProof.StorageProof[0].Proof), storageSlot, accountAddress, blockHeight)
	},
}

func init() {
	rootCmd.AddCommand(storageCmd)
	stateCmd.Flags().Int64Var(&blockHeight, "block-height", 0, "Block height (required)")
	stateCmd.Flags().StringVar(&accountAddress, "account-address", "", "Account address to inspect (required)")
	storageCmd.Flags().Int64Var(&storageSlot, "slot", 0, "Storage slot (required)")
	stateCmd.MarkFlagRequired("block-height")
	stateCmd.MarkFlagRequired("account-address")
	storageCmd.MarkFlagRequired("slot")
}
