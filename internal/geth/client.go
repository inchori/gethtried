package geth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
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

func (e *Client) GetAccountProof(ctx context.Context, address string, blockNumber int64) (*gethclient.AccountResult, error) {
	accountAddress := common.HexToAddress(address)
	blockNumBig := big.NewInt(blockNumber)

	gethClient := gethclient.New(e.ethClient.Client())
	accountProof, err := gethClient.GetProof(ctx, accountAddress, nil, blockNumBig)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof for account %s at block #%d: %v", address, blockNumber, err)
	}

	return accountProof, nil
}

func (e *Client) GetStorageProof(ctx context.Context, address string, slot int64, blockNumber int64) (*gethclient.AccountResult, error) {
	accountAddress := common.HexToAddress(address)
	blockNumBig := big.NewInt(blockNumber)

	gethClient := gethclient.New(e.ethClient.Client())

	slotBigInt := big.NewInt(slot)
	slotBytes := slotBigInt.Bytes()
	paddedSlotBytes := common.LeftPadBytes(slotBytes, 32)
	slotHex := hexutil.Encode(paddedSlotBytes)

	keys := []string{slotHex}

	storageProof, err := gethClient.GetProof(ctx, accountAddress, keys, blockNumBig)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage proof for account %s at block #%d: %v", address, blockNumber, err)
	}

	return storageProof, nil
}
