package render

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
