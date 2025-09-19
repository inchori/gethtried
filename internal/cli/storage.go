package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethtrie "github.com/ethereum/go-ethereum/trie"
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

		storageRoot := storageProof.StorageHash
		slotKey := common.LeftPadBytes(big.NewInt(storageSlot).Bytes(), 32)
		targetPathHash := crypto.Keccak256Hash(slotKey)
		targetPathNibbles := hex.EncodeToString(targetPathHash.Bytes())

		var proofBytes [][]byte
		proofMap := make(map[string]trie.RenderNodeData)
		storagePathNodes := storageProof.StorageProof[0].Proof
		var finalStorageValue []byte

		for _, nodeHexBytes := range storagePathNodes {
			rawData, _ := hexutil.Decode(nodeHexBytes)
			proofBytes = append(proofBytes, rawData)
			nodeKey := crypto.Keccak256Hash(rawData)
			parsedNode, _ := trie.ParseNode(rawData)

			proofMap[nodeKey.String()] = trie.RenderNodeData{Key: nodeKey, Node: parsedNode}

			if leaf, ok := parsedNode.(*trie.LeafNode); ok {
				finalStorageValue = leaf.Value
			}
		}

		fmt.Printf("\n--- Storage Proof Verification ---\n")

		proofDB := make(map[string][]byte)
		for _, nodeBytes := range proofBytes {
			key := crypto.Keccak256Hash(nodeBytes)
			proofDB[string(key[:])] = nodeBytes
		}

		verifiedValue, err := ethtrie.VerifyProof(storageRoot, targetPathHash.Bytes(), &MapDB{data: proofDB})
		if err != nil {
			fmt.Printf("STORAGE PROOF VERIFICATION FAILED: %v\n", err)
		} else {
			fmt.Printf("STORAGE PROOF VERIFICATION SUCCESSFUL\n")
			if len(verifiedValue) > 0 {
				fmt.Printf("   Storage Value: %s\n", hexutil.Encode(verifiedValue))
				if len(verifiedValue) == 32 {
					storageInt := new(big.Int).SetBytes(verifiedValue)
					fmt.Printf("   As Integer: %s\n", storageInt.String())
				}
			} else {
				fmt.Printf("   Storage slot is empty\n")
			}
		}

		fmt.Printf("\n--- Storage Trie Path Visualization ---\n")
		render.RenderLogicalPath(storageRoot, targetPathNibbles, proofMap, finalStorageValue)
	},
}

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.Flags().Int64Var(&storageSlot, "slot", 0, "Storage slot (required)")
	storageCmd.MarkFlagRequired("block-height")
	storageCmd.MarkFlagRequired("account-address")
	storageCmd.MarkFlagRequired("slot")
}
