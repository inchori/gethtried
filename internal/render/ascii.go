package render

import (
	"encoding/hex"
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

		childBytes := n.Children[nextNibbleIndex]
		if len(childBytes) == 0 {
			fmt.Printf("%s│   └── ERROR: Path led to an empty slot in Branch node.\n", indent)
			return
		}

		followChild(childBytes, remainingPath[1:], proofMap, finalValue, indent+"│   ")

	case *trie.ExtensionNode:
		sharedNibbles, _ := decodeHPNibbles(n.SharedPath)
		fmt.Printf("%s│   - Shared Path: '%s'\n", indent, sharedNibbles)

		if !strings.HasPrefix(remainingPath, sharedNibbles) {
			fmt.Printf("%s│   └── ERROR: Path mismatch. Expected prefix '%s' but got '%s'\n", indent, sharedNibbles, remainingPath)
			return
		}

		fmt.Printf("%s│   -> Following Extension Node...\n", indent)
		nextRemainingPath := remainingPath[len(sharedNibbles):]
		followChild(n.NextNode, nextRemainingPath, proofMap, finalValue, indent+"│   ")

	case *trie.LeafNode:
		pathEnd, _ := decodeHPNibbles(n.PathEnd)
		fmt.Printf("%s│   - Final Path: '%s'\n", indent, pathEnd)

		if remainingPath != pathEnd {
			fmt.Printf("%s│   └── ERROR: Path mismatch. Expected final path '%s' but remaining path is '%s'\n", indent, pathEnd, remainingPath)
			return
		}

		fmt.Printf("%s└── Leaf Reached. Final Value:\n", indent)
		printFinalValue(finalValue, indent+"    ")
	}
}

func followChild(childBytes []byte, remainingPath string, proofMap map[string]trie.RenderNodeData, finalValue interface{}, indent string) {
	switch len(childBytes) {
	case 32:
		childKey := hexutil.Encode(childBytes)
		walkRecursive(childKey, remainingPath, proofMap, finalValue, indent)
	case 0:
		fmt.Printf("%s└── ERROR: Empty child reference\n", indent)
	default:
		childKey := hexutil.Encode(childBytes)
		if _, ok := proofMap[childKey]; ok {
			walkRecursive(childKey, remainingPath, proofMap, finalValue, indent)
			return
		}

		hashKey := crypto.Keccak256Hash(childBytes).Hex()
		if _, ok := proofMap[hashKey]; ok {
			walkRecursive(hashKey, remainingPath, proofMap, finalValue, indent)
			return
		}

		inlineNode, err := trie.ParseNode(childBytes)
		if err != nil {
			fmt.Printf("%s└── ERROR: Failed to parse inline node: %v\n", indent, err)
			return
		}

		fmt.Printf("%s├── INLINE NODE (RLP < 32 bytes)\n", indent)
		fmt.Printf("%s│   Type: %s\n", indent, inlineNode.Type())

		walkRecursiveInline(inlineNode, remainingPath, proofMap, finalValue, indent)
	}
}

func walkRecursiveInline(node trie.Node, remainingPath string, proofMap map[string]trie.RenderNodeData, finalValue interface{}, indent string) {
	switch n := node.(type) {
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

		childBytes := n.Children[nextNibbleIndex]
		if len(childBytes) == 0 {
			fmt.Printf("%s│   └── ERROR: Path led to an empty slot in Branch node.\n", indent)
			return
		}

		followChild(childBytes, remainingPath[1:], proofMap, finalValue, indent+"│   ")

	case *trie.ExtensionNode:
		sharedNibbles, _ := decodeHPNibbles(n.SharedPath)
		fmt.Printf("%s│   - Shared Path: '%s'\n", indent, sharedNibbles)

		if !strings.HasPrefix(remainingPath, sharedNibbles) {
			fmt.Printf("%s│   └── ERROR: Path mismatch. Expected prefix '%s' but got '%s'\n", indent, sharedNibbles, remainingPath)
			return
		}

		fmt.Printf("%s│   -> Following Extension Node...\n", indent)
		nextRemainingPath := remainingPath[len(sharedNibbles):]
		followChild(n.NextNode, nextRemainingPath, proofMap, finalValue, indent+"│   ")

	case *trie.LeafNode:
		pathEnd, _ := decodeHPNibbles(n.PathEnd)
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

func decodeHPNibbles(path []byte) (string, bool) {
	if len(path) == 0 {
		return "", false
	}
	flagNibble := path[0] >> 4
	isLeaf := flagNibble == 2 || flagNibble == 3

	hexPath := hex.EncodeToString(path)
	if flagNibble%2 == 1 {
		return hexPath[1:], isLeaf
	} else {
		return hexPath[2:], isLeaf
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
