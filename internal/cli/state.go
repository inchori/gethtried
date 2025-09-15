package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/trie"
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

		proofResult, err := client.GetProof(context.Background(), accountAddress, blockHeight)
		if err != nil {
			log.Fatalf("Failed to get account proof: %v", err)
		}

		fmt.Printf("Successfully got %d proof nodes for %s at block %d.\n", len(proofResult.AccountProof), accountAddress, blockHeight)
		fmt.Println("--- Parsing Trie Nodes ---")

		for i, nodeHexStringBytes := range proofResult.AccountProof {
			rawData, err := hexutil.Decode(nodeHexStringBytes)
			if err != nil {
				log.Fatalf("Node %d: failed to decode hex string from proof: %v", i, err)
			}

			parsedNode, err := trie.ParseNode(rawData)
			if err != nil {
				log.Fatalf("Node %d: failed to parse RLP data: %v", i, err)
			}
			fmt.Printf("[Node %d] Parsed Type: %s\n", i, parsedNode.Type())
			switch node := parsedNode.(type) {
			case *trie.LeafNode:
				var accountData trie.Account
				if err := rlp.DecodeBytes(node.Value, &accountData); err != nil {
					log.Fatalf("Node %d: failed to decode account RLP data: %v", i, err)
				} else {
					fmt.Printf("   └── Leaf Node. Value (82 bytes) Decoded:\n")
					fmt.Printf("       - Nonce:       %d\n", accountData.Nonce)
					fmt.Printf("       - Balance:     %s (wei)\n", accountData.Balance.String())
					fmt.Printf("       - StorageRoot: %s\n", accountData.Root.Hex())
					fmt.Printf("       - CodeHash:    %s\n", accountData.CodeHash.Hex())
				}
			case *trie.ExtensionNode:
				fmt.Printf("   └── Extension Node. Next Node Hash: %x\n", node.NextNode)
			case *trie.BranchNode:
				fmt.Printf("   └── Branch Node. Has Value: %t\n", len(node.Value) > 0)
			}
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
