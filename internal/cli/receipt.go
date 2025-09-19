package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var receiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "Visualize the transaction receipt trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		block, err := client.GetBlockByNumber(context.Background(), blockHeight)
		if err != nil {
			log.Fatalf("Failed to get block %d: %v", blockHeight, err)
		}

		expectedRoot := block.Header().ReceiptHash

		blockReceipts, err := client.GetBlockReceipts(context.Background(), blockHeight)
		if err != nil {
			log.Fatalf("Failed to get block receipts: %v", err)
		}

		var receipts types.Receipts = blockReceipts
		fmt.Printf("Successfully fetched %d receipts for block %d.\n", len(receipts), blockHeight)

		calculatedRoot := types.DeriveSha(receipts, trie.NewStackTrie(nil))

		fmt.Printf("Block Header ReceiptRoot: %s\n", expectedRoot.Hex())
		fmt.Printf("Calculated ReceiptRoot:   %s\n", calculatedRoot.Hex())

		fmt.Println("\n--- Receipts in Trie (Key: RLP(index)) ---")
		for i, r := range receipts {
			fmt.Printf("  [Idx %d] TxHash: %s, Status: %d\n", i, r.TxHash.Hex(), r.Status)
		}
	},
}

func init() {
	rootCmd.AddCommand(receiptCmd)
	receiptCmd.MarkFlagRequired("block-height")
}
