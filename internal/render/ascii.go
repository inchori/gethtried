package render

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/inchori/geth-state-trie/internal/trie"
)

func RenderProofPath(pathNodes []trie.Node, value interface{}) {
	fmt.Println("--- ASCII Trie Path Visualization ---")

	indent := "  "
	for i, node := range pathNodes {
		prefix := "├──"
		if i == len(pathNodes)-1 { // 마지막 노드면
			prefix = "└──"
		}

		fmt.Printf("%s%s [Node %d] Type: %s\n", indent, prefix, i, node.Type())
		indent += "│   "

		switch n := node.(type) {
		case *trie.LeafNode:
			switch val := value.(type) {
			case *trie.Account:
				weiFloat := new(big.Float).SetInt(val.Balance)
				ethConstantFloat := new(big.Float).SetInt64(params.Ether)
				ethValue := new(big.Float).Quo(weiFloat, ethConstantFloat)

				fmt.Printf("%s   - Nonce:       %d\n", indent, val.Nonce)
				fmt.Printf("%s   - Balance:     %s ETH\n", indent, ethValue.Text('f', 6))
				fmt.Printf("%s   - StorageRoot: %s\n", indent, val.Root.Hex())
				fmt.Printf("%s   - CodeHash:    %s\n", indent, val.CodeHash.Hex())
			case []byte:
				fmt.Printf("%s   - Value (32 bytes): %s\n", indent, hexutil.Encode(val))
			default:
				fmt.Printf("%s   -  Raw Value: %v\n", indent, n.Value)
			}
		case *trie.ExtensionNode:
			fmt.Printf("%s   - Next Hash: %x\n", indent, n.NextNode)
		case *trie.BranchNode:
			fmt.Printf("%s   - Has Value: %t\n", indent, len(n.Value) > 0)
		}
	}
}
