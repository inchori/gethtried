package geth_test

import (
	"context"
	"fmt"
	"geth-state-trie/geth"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProof(t *testing.T) {
	rpcUrl := ""
	ethClient, err := geth.NewEthClient(rpcUrl)
	assert.NoError(t, err)

	proof, err := ethClient.GetProof(context.TODO(), "0xC2f6b569dE849f59f49709cA4A0c5f682AC70241",
		23340312)
	assert.NoError(t, err)
	fmt.Println(proof)
}
