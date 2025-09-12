package geth

import (
	"context"
	"fmt"
	"math/big"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	ethClient *ethclient.Client
}

func NewEthClient(rpcUrl string) (*Client, error) {
	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to dial geth rpc: %v", err)
	}
	return &Client{ethClient: ethClient}, nil
}

func (e *Client) GetBlockByNumber(ctx context.Context, blockHeight int64) (*gethtypes.Block, error) {
	block, err := e.ethClient.BlockByNumber(ctx, big.NewInt(blockHeight))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %v", err)
	}

	return block, nil
}
