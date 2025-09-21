package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/gethtried/internal/geth"
	"github.com/inchori/gethtried/internal/render"
	"github.com/inchori/gethtried/internal/trie"
	"github.com/spf13/cobra"
)

var storageSlotStr string

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Visualize the storage trie for a specific account at a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStorageCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runStorageCommand() error {
	if !common.IsHexAddress(accountAddress) {
		return fmt.Errorf("invalid account address format: %s (expected format: 0x...)", accountAddress)
	}

	if blockHeight < 0 {
		return fmt.Errorf("block height must be non-negative, got: %d", blockHeight)
	}

	var storageSlot int64
	if strings.HasPrefix(storageSlotStr, "0x") || strings.HasPrefix(storageSlotStr, "0X") {
		slotBig, ok := new(big.Int).SetString(storageSlotStr[2:], 16)
		if !ok {
			return fmt.Errorf("invalid hex storage slot: %s", storageSlotStr)
		}
		if !slotBig.IsInt64() {
			return fmt.Errorf("storage slot too large: %s (max: %d)", storageSlotStr, int64(^uint64(0)>>1))
		}
		storageSlot = slotBig.Int64()
	} else {
		var err error
		storageSlot, err = strconv.ParseInt(storageSlotStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid storage slot: %s (must be decimal number or hex with 0x prefix)", storageSlotStr)
		}
	}

	if storageSlot < 0 {
		return fmt.Errorf("storage slot must be non-negative, got: %d", storageSlot)
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

	storageProof, err := client.GetStorageProof(context.Background(), accountAddress, storageSlot, blockHeight)
	if err != nil {
		return fmt.Errorf("failed to get storage proof for %s slot %d at block %d: %w", accountAddress, storageSlot, blockHeight, err)
	}

	if len(storageProof.StorageProof) == 0 {
		return fmt.Errorf("no storage proof returned for slot %d (slot may not exist)", storageSlot)
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

	return nil
}

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.Flags().StringVar(&storageSlotStr, "slot", "0", "Storage slot (decimal or hex with 0x prefix)")
	storageCmd.MarkFlagRequired("block-height")
	storageCmd.MarkFlagRequired("account-address")
	storageCmd.MarkFlagRequired("slot")
}
