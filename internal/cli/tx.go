package cli

import (
	"context"
	"fmt"
	"log"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/spf13/cobra"
)

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Visualize the transaction trie for a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		block, err := client.GetBlockByNumber(context.Background(), blockHeight)
		if err != nil {
			log.Fatalf("Failed to get block by height: %v", err)
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
	},
}

func init() {
	rootCmd.AddCommand(txCmd)
	txCmd.MarkFlagRequired("block-height")
}
