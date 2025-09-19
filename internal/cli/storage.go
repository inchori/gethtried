package cli

import (
	"context"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

		startRootKey := storageProof.StorageHash

		slotKey := common.LeftPadBytes(big.NewInt(storageSlot).Bytes(), 32)
		targetPathHash := crypto.Keccak256Hash(slotKey)
		targetPathNibbles := hex.EncodeToString(targetPathHash.Bytes())

		proofMap := make(map[string]trie.RenderNodeData)
		storagePathNodes := storageProof.StorageProof[0].Proof
		var finalStorageValue []byte

		for _, nodeHexBytes := range storagePathNodes {
			rawData, _ := hexutil.Decode(nodeHexBytes)
			nodeKey := crypto.Keccak256Hash(rawData)
			parsedNode, _ := trie.ParseNode(rawData)

			proofMap[nodeKey.String()] = trie.RenderNodeData{Key: nodeKey, Node: parsedNode}

			if leaf, ok := parsedNode.(*trie.LeafNode); ok {
				finalStorageValue = leaf.Value
			}
		}
		render.RenderLogicalPath(startRootKey, targetPathNibbles, proofMap, finalStorageValue)
	},
}

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.Flags().Int64Var(&storageSlot, "slot", 0, "Storage slot (required)")
	storageCmd.MarkFlagRequired("block-height")
	storageCmd.MarkFlagRequired("account-address")
	storageCmd.MarkFlagRequired("slot")
}
