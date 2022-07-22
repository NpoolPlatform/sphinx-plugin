package bsc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MinNodeNum = 1
	MaxRetries = 3
)

type BClientI interface {
	BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error)
	NetworkIDS(ctx context.Context) (*big.Int, error)
	SuggestGasPriceS(ctx context.Context) (*big.Int, error)
	SendTransactionS(ctx context.Context, tx *types.Transaction) error
	TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*ethclient.Client, error)
	WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error
}

type bClients struct{}

func (bClients bClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*ethclient.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}
	cli, err := ethclient.DialContext(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	syncRet, _err := cli.SyncProgress(ctx)
	if _err != nil {
		cli.Close()
		return nil, _err
	}

	if syncRet != nil {
		cli.Close()
		return nil, fmt.Errorf(
			"node is syncing ,current block %v ,highest block %v ",
			syncRet.CurrentBlock, syncRet.HighestBlock,
		)
	}

	return cli, nil
}

func (bClients *bClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
	var (
		apiErr, err error
		retry       bool
	)
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}
	for i := 0; i < MaxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		client, err := bClients.GetNode(ctx, endpointmgr)
		if errors.Is(err, endpoints.ErrEndpointExhausted) {
			if apiErr != nil {
				return apiErr
			}
			return err
		}
		if err != nil {
			continue
		}

		retry, apiErr = fn(ctx, client)
		client.Close()

		if apiErr != nil || !retry {
			return apiErr
		}
	}
	return err
}

func (bClients bClients) BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.BalanceAt(ctx, account, blockNumber)
		return false, err
	})
	return ret, err
}

func (bClients bClients) PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error) {
	var ret uint64
	var err error

	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.PendingNonceAt(ctx, account)
		return false, err
	})
	return ret, err
}

func (bClients bClients) NetworkIDS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.NetworkID(ctx)
		return false, err
	})

	return ret, err
}

func (bClients bClients) SuggestGasPriceS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.SuggestGasPrice(ctx)
		return false, err
	})

	return ret, err
}

func (bClients bClients) SendTransactionS(ctx context.Context, tx *types.Transaction) error {
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		err = c.SendTransaction(ctx, tx)
		if err != nil && TxFailErr(err) {
			return false, err
		}
		return false, err
	})

	return err
}

func (bClients bClients) TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		tx, isPending, err = c.TransactionByHash(ctx, hash)
		return false, err
	})

	return tx, isPending, err
}

func (bClients bClients) TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var ret *types.Receipt
	var err error
	err = bClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.TransactionReceipt(ctx, txHash)
		return false, err
	})

	return ret, err
}

func Client() BClientI {
	return &bClients{}
}
