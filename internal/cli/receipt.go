package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var receiptCmd = &cobra.Command{
	Use:   "receipt",
	Short: "Visualize the transaction receipt trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runReceiptCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runReceiptCommand() error {
	if blockHeight < 0 {
		return fmt.Errorf("block height must be non-negative, got: %d", blockHeight)
	}

	client, err := geth.NewEthClient(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC endpoint %s: %w", rpcURL, err)
	}

	latestBlock, err := client.GetBlockByNumber(context.Background(), -1)
	if err != nil {
		return fmt.Errorf("failed to get latest block (check RPC connection): %w", err)
	}

	if uint64(blockHeight) > latestBlock.NumberU64() {
		return fmt.Errorf("block height %d exceeds latest block %d", blockHeight, latestBlock.NumberU64())
	}

	block, err := client.GetBlockByNumber(context.Background(), blockHeight)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %w", blockHeight, err)
	}

	expectedRoot := block.Header().ReceiptHash

	blockReceipts, err := client.GetBlockReceipts(context.Background(), blockHeight)
	if err != nil {
		return fmt.Errorf("failed to get block receipts for block %d: %w", blockHeight, err)
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

	return nil
}

func init() {
	rootCmd.AddCommand(receiptCmd)
	receiptCmd.MarkFlagRequired("block-height")
}
