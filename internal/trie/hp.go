package trie

import "encoding/hex"

func DecodeHP(path []byte) (string, bool) {
	if len(path) == 0 {
		return "", false
	}

	firstByte := path[0]
	firstNibble := firstByte >> 4
	isLeaf := firstNibble == 2 || firstNibble == 3

	nibbles := ""
	if firstNibble%2 == 1 {
		nibbles = string(hex.EncodeToString([]byte{firstByte & 0x0F}))
		path = path[1:]
	} else {
		nibbles = ""
		path = path[1:]
	}

	nibbles += hex.EncodeToString(path)
	return nibbles, isLeaf
}
