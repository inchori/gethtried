package cli

import (
	"context"
	"fmt"
	"os"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/gethtried/internal/geth"
	"github.com/spf13/cobra"
)

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Visualize the transaction trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTxCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runTxCommand() error {
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
	expectedRoot := block.Header().TxHash
	transactions := block.Transactions()

	calculatedRoot := gethtypes.DeriveSha(transactions, gethtrie.NewStackTrie(nil))

	fmt.Printf("Block Header TxRoot: %s\n", expectedRoot.Hex())
	fmt.Printf("Calculated TxRoot:   %s\n", calculatedRoot.Hex())
	if expectedRoot == calculatedRoot {
		fmt.Println("Verification Successful!")
	} else {
		fmt.Println("Verification FAILED!")
	}

	fmt.Println("\n--- Transactions in Trie (Key: RLP(index)) ---")
	for i, tx := range transactions {
		fmt.Printf("  [Idx %d] TxHash: %s\n", i, tx.Hash().Hex())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(txCmd)
	txCmd.MarkFlagRequired("block-height")
}
