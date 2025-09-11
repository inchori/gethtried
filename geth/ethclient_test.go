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

	proof, err := ethClient.GetProof(context.TODO(), "",
		23340312)
	assert.NoError(t, err)
	fmt.Println(proof)
}
