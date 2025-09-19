package trie

import "github.com/ethereum/go-ethereum/common"

type Node interface {
	Type() string
}

type BranchNode struct {
	Children [16][]byte
	Value    []byte
}

type ExtensionNode struct {
	SharedPath []byte
	NextNode   []byte
}

type LeafNode struct {
	PathEnd []byte
	Value   []byte
}

type RenderNode struct {
	Key  []byte
	Node Node
}

func (b *BranchNode) Type() string { return "Branch" }

func (e *ExtensionNode) Type() string { return "Extension" }

func (l *LeafNode) Type() string { return "Leaf" }

type RenderNodeData struct {
	Key  common.Hash
	Node Node
}
