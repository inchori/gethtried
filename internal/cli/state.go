package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	"github.com/inchori/geth-state-trie/internal/geth"
	"github.com/inchori/geth-state-trie/internal/render"
	"github.com/inchori/geth-state-trie/internal/trie"
	"github.com/spf13/cobra"
)

type MapDB struct {
	data map[string][]byte
}

func (db *MapDB) Get(key []byte) ([]byte, error) {
	if value, ok := db.data[string(key)]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (db *MapDB) Has(key []byte) (bool, error) {
	_, ok := db.data[string(key)]
	return ok, nil
}

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Visualize the state trie for a specific account at a specific block height",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStateCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runStateCommand() error {
	if !common.IsHexAddress(accountAddress) {
		return fmt.Errorf("invalid account address format: %s (expected format: 0x...)", accountAddress)
	}

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
	stateRoot := block.Header().Root

	proofResult, err := client.GetAccountProof(context.Background(), accountAddress, blockHeight)
	if err != nil {
		return fmt.Errorf("failed to get account proof for %s at block %d: %w", accountAddress, blockHeight, err)
	}

	fmt.Printf("Successfully got %d proof nodes for %s at block %d.\n", len(proofResult.AccountProof), accountAddress, blockHeight)

	var proofBytes [][]byte
	var renderNodeList []trie.RenderNodeData
	var finalValue interface{}

	for i, nodeHexStringBytes := range proofResult.AccountProof {
		rawData, err := hexutil.Decode(nodeHexStringBytes)
		if err != nil {
			return fmt.Errorf("failed to decode proof node %d: %w", i, err)
		}

		proofBytes = append(proofBytes, rawData)
		nodeKey := crypto.Keccak256Hash(rawData)

		parsedNode, err := trie.ParseNode(rawData)
		if err != nil {
			return fmt.Errorf("failed to parse proof node %d: %w", i, err)
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

	fmt.Printf("\n--- Cryptographic Proof Verification ---\n")

	proofDB := make(map[string][]byte)
	for _, nodeBytes := range proofBytes {
		key := crypto.Keccak256Hash(nodeBytes)
		proofDB[string(key[:])] = nodeBytes
	}

	verifiedValue, err := ethtrie.VerifyProof(stateRoot, targetPathHash, &MapDB{data: proofDB})
	if err != nil {
		fmt.Printf("PROOF VERIFICATION FAILED: %v\n", err)
	} else {
		fmt.Printf("PROOF VERIFICATION SUCCESSFUL\n")

		if len(verifiedValue) > 0 {
			var verifiedAccount trie.Account
			if err := rlp.DecodeBytes(verifiedValue, &verifiedAccount); err == nil {
				fmt.Printf("   Verified Account Data:\n")
				fmt.Printf("   - Nonce: %d\n", verifiedAccount.Nonce)
				fmt.Printf("   - Balance: %s wei\n", verifiedAccount.Balance.String())
				fmt.Printf("   - Storage Root: %s\n", verifiedAccount.Root.Hex())
				fmt.Printf("   - Code Hash: %s\n", verifiedAccount.CodeHash.Hex())
			} else {
				fmt.Printf("   Raw verified value: %s\n", hexutil.Encode(verifiedValue))
			}
		} else {
			fmt.Printf("   Account does not exist (empty proof)\n")
		}
	}

	proofMap := make(map[string]trie.RenderNodeData)
	for _, rn := range renderNodeList {
		proofMap[rn.Key.String()] = rn
	}

	startRootKey := renderNodeList[0].Key
	fmt.Printf("\n--- Trie Path Visualization ---\n")
	render.RenderLogicalPath(startRootKey, targetPathNibbles, proofMap, finalValue)

	return nil
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.MarkFlagRequired("block-height")
	stateCmd.MarkFlagRequired("account-address")
}
