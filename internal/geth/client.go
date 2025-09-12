package geth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient struct {
	gethRpcClient *gethclient.Client
}

func NewEthClient(rpcUrl string) (*EthClient, error) {
	client, err := rpc.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to dial geth rpc: %v", err)
	}

	gethRpcClient := gethclient.New(client)
	return &EthClient{gethRpcClient: gethRpcClient}, nil
}

func (e *EthClient) GetProof(ctx context.Context, address string, blockNumber int64) (*gethclient.AccountResult, error) {
	commonAddress := common.HexToAddress(address)
	intBlockNumber := big.NewInt(blockNumber)

	proof, err := e.gethRpcClient.GetProof(ctx, commonAddress, nil, intBlockNumber)
	if err != nil {
		return nil, err
	}

	return proof, nil
}
