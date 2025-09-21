package render

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/inchori/geth-state-trie/internal/trie"
)

func RenderProofPath(pathNodes []trie.RenderNode, value interface{}) {
	fmt.Println("--- ASCII Trie Path Visualization ---")

	indent := "  "
	for i, data := range pathNodes {
		prefix := "├──"
		if i == len(pathNodes)-1 {
			prefix = "└──"
		}

		fmt.Printf("%s%s [Node %d] KEY: %s\n", indent, prefix, i, hexutil.Encode(data.Key))

		detailsIndent := indent + "│   "
		fmt.Printf("%s     Type: %s\n", detailsIndent, data.Node.Type())

		switch n := data.Node.(type) {
		case *trie.LeafNode:
			switch val := value.(type) {
			case *trie.Account:
				weiFloat := new(big.Float).SetInt(val.Balance)
				ethConstantFloat := new(big.Float).SetInt64(params.Ether)
				ethValue := new(big.Float).Quo(weiFloat, ethConstantFloat)

				fmt.Printf("%s   - Nonce:       %d\n", detailsIndent, val.Nonce)
				fmt.Printf("%s   - Balance:     %s ETH\n", detailsIndent, ethValue.Text('f', 6))
				fmt.Printf("%s   - StorageRoot: %s\n", detailsIndent, val.Root.Hex())
				fmt.Printf("%s   - CodeHash:    %s\n", detailsIndent, val.CodeHash.Hex())
			case []byte:
				fmt.Printf("%s   - Value (32 bytes): %s\n", detailsIndent, hexutil.Encode(val))
			default:
				fmt.Printf("%s   -  Raw Value: %v\n", detailsIndent, n.Value)
			}
		case *trie.ExtensionNode:
			fmt.Printf("%s   - Next Hash: %x\n", detailsIndent, n.NextNode)
		case *trie.BranchNode:
			fmt.Printf("%s   - Has Value: %t\n", detailsIndent, len(n.Value) > 0)
		}
	}
}

func RenderLogicalPath(
	startRootKey common.Hash,
	targetPathNibbles string,
	proofMap map[string]trie.RenderNodeData,
	finalValue interface{},
) {
	fmt.Println("--- Logical Trie Path Visualization ---")
	fmt.Printf("Target Path: %s\n", targetPathNibbles)

	walkRecursive(startRootKey.Hex(), targetPathNibbles, proofMap, finalValue, "")
}

func walkRecursive(
	currentNodeKey string,
	remainingPath string,
	proofMap map[string]trie.RenderNodeData,
	finalValue interface{},
	indent string,
) {
	data, ok := proofMap[currentNodeKey]
	if !ok {
		fmt.Printf("%s└── ERROR: Missing node in proof path! Hash: %s\n", indent, currentNodeKey)
		return
	}

	fmt.Printf("%s├── KEY: %s\n", indent, currentNodeKey)
	fmt.Printf("%s│   Type: %s\n", indent, data.Node.Type())

	switch n := data.Node.(type) {

	case *trie.BranchNode:
		fmt.Printf("%s│   - Has Value: %t\n", indent, len(n.Value) > 0)

		if len(remainingPath) == 0 {
			if len(n.Value) > 0 {
				fmt.Printf("%s└── Branch value reached. Final Value:\n", indent)
				printFinalValue(finalValue, indent+"    ")
				return
			}
			fmt.Printf("%s│   └── ERROR: Path ended at Branch without value\n", indent)
			return
		}

		nextNibbleChar := remainingPath[0]
		nextNibbleIndex := hexNibbleToIndex(nextNibbleChar)
		if nextNibbleIndex == -1 {
			fmt.Printf("%s│   └── ERROR: Invalid path nibble '%c'\n", indent, nextNibbleChar)
			return
		}

		fmt.Printf("%s│   -> Branching: Following path nibble '%c' (index %d)\n", indent, nextNibbleChar, nextNibbleIndex)

		childRef := n.Children[nextNibbleIndex]
		if len(childRef) == 0 {
			fmt.Printf("%s│   └── ERROR: Path led to an empty slot in Branch node.\n", indent)
			return
		}

		var childKeyHex string
		if len(childRef) == 32 {
			childKeyHex = hexutil.Encode(childRef)
		} else {
			h := crypto.Keccak256(childRef)
			childKeyHex = hexutil.Encode(h)
			parsedChild, err := trie.ParseNode(childRef)
			if err != nil {
				fmt.Printf("%s│   └── ERROR: Failed to parse inline child node: %v\n", indent, err)
				return
			}
			proofMap[common.BytesToHash(h).Hex()] = trie.RenderNodeData{Key: common.BytesToHash(h), Node: parsedChild}
		}

		walkRecursive(childKeyHex, remainingPath[1:], proofMap, finalValue, indent+"│   ")

	case *trie.ExtensionNode:
		sharedNibbles, _ := trie.DecodeHP(n.SharedPath)
		fmt.Printf("%s│   - Shared Path: '%s'\n", indent, sharedNibbles)

		if !strings.HasPrefix(remainingPath, sharedNibbles) {
			fmt.Printf("%s│   └── ERROR: Path mismatch. Expected prefix '%s' but got '%s'\n", indent, sharedNibbles, remainingPath)
			return
		}

		fmt.Printf("%s│   -> Following Extension Node...\n", indent)
		nextRef := n.NextNode
		nextRemainingPath := remainingPath[len(sharedNibbles):]

		var nextKeyHex string
		if len(nextRef) == 32 {
			nextKeyHex = hexutil.Encode(nextRef)
		} else {
			h := crypto.Keccak256(nextRef)
			nextKeyHex = hexutil.Encode(h)
			parsedNext, err := trie.ParseNode(nextRef)
			if err != nil {
				fmt.Printf("%s│   └── ERROR: Failed to parse inline next node: %v\n", indent, err)
				return
			}
			proofMap[common.BytesToHash(h).Hex()] = trie.RenderNodeData{Key: common.BytesToHash(h), Node: parsedNext}
		}

		walkRecursive(nextKeyHex, nextRemainingPath, proofMap, finalValue, indent+"│   ")

	case *trie.LeafNode:
		pathEnd, _ := trie.DecodeHP(n.PathEnd)
		fmt.Printf("%s│   - Final Path: '%s'\n", indent, pathEnd)

		if remainingPath != pathEnd {
			fmt.Printf("%s│   └── ERROR: Path mismatch. Expected final path '%s' but remaining path is '%s'\n", indent, pathEnd, remainingPath)
			return
		}

		fmt.Printf("%s└── Leaf Reached. Final Value:\n", indent)
		printFinalValue(finalValue, indent+"    ")
	}
}

func printFinalValue(finalValue interface{}, indent string) {
	switch val := finalValue.(type) {
	case *trie.Account:
		weiFloat := new(big.Float).SetInt(val.Balance)
		ethConstantFloat := new(big.Float).SetInt64(params.Ether)
		ethValue := new(big.Float).Quo(weiFloat, ethConstantFloat)

		fmt.Printf("%s- Nonce:       %d\n", indent, val.Nonce)
		fmt.Printf("%s- Balance:     %s ETH\n", indent, ethValue.Text('f', 6))
		fmt.Printf("%s- StorageRoot: %s\n", indent, val.Root.Hex())
		fmt.Printf("%s- CodeHash:    %s\n", indent, val.CodeHash.Hex())

	case []byte:
		fmt.Printf("%s- Value: %s\n", indent, hexutil.Encode(val))

	default:
		fmt.Printf("%s- Unknown Value Type\n", indent)
	}
}

func hexNibbleToIndex(nibbleChar byte) int {
	if nibbleChar >= '0' && nibbleChar <= '9' {
		return int(nibbleChar - '0')
	}
	if nibbleChar >= 'a' && nibbleChar <= 'f' {
		return int(nibbleChar-'a') + 10
	}
	if nibbleChar >= 'A' && nibbleChar <= 'F' {
		return int(nibbleChar-'A') + 10
	}
	log.Fatalf("Invalid hex nibble character: %c", nibbleChar)
	return -1
}
