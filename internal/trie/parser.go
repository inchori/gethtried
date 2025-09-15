package trie

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

func ParseNode(rawData []byte) (Node, error) {
	if len(rawData) == 0 {
		return nil, fmt.Errorf("empty raw data")
	}

	var decodedList [][]byte
	if err := rlp.DecodeBytes(rawData, &decodedList); err != nil {
		return nil, fmt.Errorf("failed to decode RLP: %v", err)
	}

	switch len(decodedList) {
	case 17:
		var children [16][]byte
		copy(children[:], decodedList[:16])

		return &BranchNode{
			Children: children,
			Value:    decodedList[16],
		}, nil
	case 2:
		pathBytes := decodedList[0]
		valueOrHash := decodedList[1]

		if len(pathBytes) == 0 {
			return nil, fmt.Errorf("node has 2 items but path is empty")
		}

		firstNibble := pathBytes[0] >> 4
		if firstNibble == 0 || firstNibble == 1 {
			return &ExtensionNode{
				SharedPath: pathBytes,
				NextNode:   valueOrHash,
			}, nil
		} else if firstNibble == 2 || firstNibble == 3 {
			return &LeafNode{
				PathEnd: pathBytes,
				Value:   valueOrHash,
			}, nil
		} else {
			return nil, fmt.Errorf("invalid hex-prefix nibble: %x", firstNibble)
		}
	default:
		return nil, fmt.Errorf("invalid node with %d items", len(decodedList))
	}
}
