package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/render"
	"github.com/inchori/geth-state-trie/internal/trie"
	"github.com/spf13/cobra"
)

var storageSlotStr string

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Visualize the storage trie for a specific account at a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		if !common.IsHexAddress(accountAddress) {
			log.Fatalf("Invalid account address format: %s", accountAddress)
		}

		if blockHeight < 0 {
			log.Fatalf("Block height must be non-negative: %d", blockHeight)
		}

		var storageSlot int64
		if strings.HasPrefix(storageSlotStr, "0x") || strings.HasPrefix(storageSlotStr, "0X") {
			slotBig, ok := new(big.Int).SetString(storageSlotStr[2:], 16)
			if !ok {
				log.Fatalf("Invalid hex storage slot: %s", storageSlotStr)
			}
			if !slotBig.IsInt64() {
				log.Fatalf("Storage slot too large: %s", storageSlotStr)
			}
			storageSlot = slotBig.Int64()
		} else {
			var err error
			storageSlot, err = strconv.ParseInt(storageSlotStr, 10, 64)
			if err != nil {
				log.Fatalf("Invalid storage slot: %s", storageSlotStr)
			}
		}

		if storageSlot < 0 {
			log.Fatalf("Storage slot must be non-negative: %d", storageSlot)
		}

		client, err := geth.NewEthClient(rpcURL)
		if err != nil {
			log.Fatalf("Failed to initialize eth client: %v", err)
		}

		latestBlock, err := client.GetBlockByNumber(context.Background(), -1)
		if err != nil {
			log.Fatalf("Failed to get latest block: %v", err)
		}

		if uint64(blockHeight) > latestBlock.NumberU64() {
			log.Fatalf("Block height %d exceeds latest block %d", blockHeight, latestBlock.NumberU64())
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
	storageCmd.Flags().StringVar(&storageSlotStr, "slot", "0", "Storage slot (decimal or hex with 0x prefix)")
	storageCmd.MarkFlagRequired("block-height")
	storageCmd.MarkFlagRequired("account-address")
	storageCmd.MarkFlagRequired("slot")
}
