package trie

import "encoding/hex"

func DecodeHP(path []byte) (string, bool) {
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
