package bsc

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Init() {
	rand.Seed(time.Now().Unix())
}

const (
	MinNodeNum = 1
	MaxRetries = 3
)

var (
	ErrGasToLow   = "intrinsic gas too low"
	ErrFundsToLow = "insufficient funds for gas * price + value"
)

type BClientI interface {
	BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error)
	NetworkIDS(ctx context.Context) (*big.Int, error)
	SuggestGasPriceS(ctx context.Context) (*big.Int, error)
	SendTransactionS(ctx context.Context, tx *types.Transaction) error
	TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetNode(mustLocalEndpoint bool) (*ethclient.Client, error)
	WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error
}

type BClients struct{}

func (bClients BClients) GetNode(mustLocalEndpoint bool) (*ethclient.Client, error) {
	addr, _, err := endpoints.Peek(mustLocalEndpoint)
	if err != nil {
		return nil, err
	}
	return ethclient.Dial(addr)
}

func (bClients *BClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
	var client *ethclient.Client
	var err error
	var retry bool
	mustLocal := true
	for i := 0; i < MaxRetries; i++ {
		client, err = bClients.GetNode(mustLocal)
		mustLocal = false
		if err != nil {
			continue
		}
		defer client.Close()
		retry, err = fn(ctx, client)
		if err == nil || !retry {
			return err
		}
	}
	return err
}

func (bClients BClients) BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		syncRet, syncErr := c.SyncProgress(ctx)
		if syncErr != nil {
			return true, syncErr
		}
		if syncRet != nil && syncRet.CurrentBlock < syncRet.HighestBlock {
			return true, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.CurrentBlock, syncRet.HighestBlock,
			)
		}
		ret, err = c.BalanceAt(ctx, account, blockNumber)
		return true, err
	})
	if err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("fail BlanceAtS, %v", err)
}

func (bClients BClients) PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error) {
	var ret uint64
	var err error

	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.PendingNonceAt(ctx, account)
		return true, err
	})

	if err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("fail PendingNonceAtS, %v", err)
}

func (bClients BClients) NetworkIDS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.NetworkID(ctx)
		return true, err
	})
	if err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("fail NetworkIDS, %v", err)
}

func (bClients BClients) SuggestGasPriceS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.SuggestGasPrice(ctx)
		return true, err
	})
	if err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("fail SuggestGasPriceS, %v", err)
}

func (bClients BClients) SendTransactionS(ctx context.Context, tx *types.Transaction) error {
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		err = c.SendTransaction(ctx, tx)
		if strings.Contains(err.Error(), ErrFundsToLow) || strings.Contains(err.Error(), ErrGasToLow) {
			return false, err
		}
		return true, err
	})

	return fmt.Errorf("fail SendTransactionS, %v", err)
}

func (bClients BClients) TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		tx, isPending, err = c.TransactionByHash(ctx, hash)
		return true, err
	})
	if err == nil {
		return tx, isPending, nil
	}
	return tx, isPending, fmt.Errorf("fail TransactionByHashS, %v", err)
}

func (bClients BClients) TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var ret *types.Receipt
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.TransactionReceipt(ctx, txHash)
		return true, err
	})
	if err == nil {
		return ret, nil
	}
	return ret, fmt.Errorf("fail TransactionReceiptS, %v", err)
}

func Client() BClientI {
	return &BClients{}
}
