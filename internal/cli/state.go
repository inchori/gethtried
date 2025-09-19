package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/render"
	"github.com/inchori/geth-state-trie/internal/trie"
	"github.com/spf13/cobra"
)

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

		//var parsedNodeList []trie.Node
		var renderNodeList []trie.RenderNodeData
		//var finalAccountData *trie.Account

		var finalValue interface{}

		for i, nodeHexStringBytes := range proofResult.AccountProof {
			rawData, err := hexutil.Decode(nodeHexStringBytes)
			if err != nil {
				log.Fatalf("Node %d: failed to decode hex string from proof: %v", i, err)
			}

			nodeKey := crypto.Keccak256Hash(rawData)

			parsedNode, err := trie.ParseNode(rawData)
			if err != nil {
				log.Fatalf("Failed to parse RLP data: %v", err)
			}

			renderNodeList = append(renderNodeList, trie.RenderNodeData{
				Key:  nodeKey,
				Node: parsedNode,
			})

			if leafNode, ok := parsedNode.(*trie.LeafNode); ok {
				var accountData trie.Account
				if err := rlp.DecodeBytes(leafNode.Value, &accountData); err == nil {
					finalValue = &accountData
				}
			}
		}

		targetPathHash := crypto.Keccak256(common.HexToAddress(accountAddress).Bytes())
		targetPathNibbles := hex.EncodeToString(targetPathHash)

		proofMap := make(map[string]trie.RenderNodeData)
		for _, rn := range renderNodeList {
			proofMap[rn.Key.String()] = rn
		}

		startRootKey := renderNodeList[0].Key
		//render.RenderProofPath(renderNodeList, finalValue)
		render.RenderLogicalPath(startRootKey, targetPathNibbles, proofMap, finalValue)
	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
}
