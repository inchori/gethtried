package trie

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash
	CodeHash common.Hash
}
