package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/render"
	"github.com/inchori/geth-state-trie/internal/trie"
	"github.com/spf13/cobra"
)

var blockHeight int64
var accountAddress string

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Visualize the state trie for a specific account at a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		proofResult, err := client.GetAccountProof(context.Background(), accountAddress, blockHeight)
		if err != nil {
			log.Fatalf("Failed to get account proof: %v", err)
		}

		fmt.Printf("Successfully got %d proof nodes for %s at block %d.\n", len(proofResult.AccountProof), accountAddress, blockHeight)

		var parsedNodeList []trie.Node
		var finalAccountData *trie.Account

		for i, nodeHexStringBytes := range proofResult.AccountProof {
			rawData, err := hexutil.Decode(nodeHexStringBytes)
			if err != nil {
				log.Fatalf("Node %d: failed to decode hex string from proof: %v", i, err)
			}

			parsedNode, err := trie.ParseNode(rawData)
			if err != nil {
				log.Fatalf("Node %d: failed to parse RLP data: %v", i, err)
			}

			parsedNodeList = append(parsedNodeList, parsedNode)

			if leafNode, ok := parsedNode.(*trie.LeafNode); ok {
				var accountData trie.Account
				if err := rlp.DecodeBytes(leafNode.Value, &accountData); err == nil {
					finalAccountData = &accountData
				}
			}
		}
		render.RenderProofPath(parsedNodeList, finalAccountData)
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.Flags().Int64Var(&blockHeight, "block-height", 0, "Block height (required)")
	stateCmd.Flags().StringVar(&accountAddress, "account-address", "", "Account address to inspect (required)")
	stateCmd.MarkFlagRequired("block-height")
	stateCmd.MarkFlagRequired("account-address")
}
