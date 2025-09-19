package cli

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/render"
	"github.com/inchori/geth-state-trie/internal/trie"
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

		if len(storageProof.StorageProof) == 0 {
			log.Fatalf("No storage proof returned for slot %d", storageSlot)
		}

		log.Printf("Successfully got %d proof nodes for storage slot %d of account %s at block %d.\n", len(storageProof.StorageProof[0].Proof), storageSlot, accountAddress, blockHeight)

		//var parsedNodeList []trie.Node
		var renderNodeList []trie.RenderNode
		//var finalStorageValue []byte
		var finalValue interface{}

		storageProofPath := storageProof.StorageProof[0].Proof

		for i, nodeHexStringBytes := range storageProofPath {
			rawData, err := hexutil.Decode(nodeHexStringBytes)
			if err != nil {
				log.Fatalf("Node %d: failed to decode hex string from proof: %v", i, err)
			}

			nodeKey := crypto.Keccak256(rawData)

			parsedNode, err := trie.ParseNode(rawData)
			if err != nil {
				log.Fatalf("Node %d: failed to parse RLP data: %v", i, err)
			}

			renderNodeList = append(renderNodeList, trie.RenderNode{
				Key:  nodeKey,
				Node: parsedNode,
			})

			if leafNode, ok := parsedNode.(*trie.LeafNode); ok {
				finalValue = leafNode.Value
			}
		}
		render.RenderProofPath(renderNodeList, finalValue)
	},
}

func init() {
	rootCmd.AddCommand(storageCmd)

	storageCmd.Flags().Int64Var(&storageSlot, "slot", 0, "Storage slot (required)")
	storageCmd.MarkFlagRequired("slot")
}
