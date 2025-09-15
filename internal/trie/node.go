package trie

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

func (b *BranchNode) Type() string { return "Branch" }

func (e *ExtensionNode) Type() string { return "Extension" }

func (l *LeafNode) Type() string { return "Leaf" }
