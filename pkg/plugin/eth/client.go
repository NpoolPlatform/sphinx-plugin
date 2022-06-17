package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MinNodeNum = 1
	MaxRetries = 3
)

var (
	ErrGasToLow   = "intrinsic gas too low"
	ErrFundsToLow = "insufficient funds for gas * price + value"
)

type EClientI interface {
	BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error)
	NetworkIDS(ctx context.Context) (*big.Int, error)
	SuggestGasPriceS(ctx context.Context) (*big.Int, error)
	SendTransactionS(ctx context.Context, tx *types.Transaction) error
	TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetNode(localEndpoint bool) (*ethclient.Client, error)
	WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error
}

type EClients struct{}

func (eClients EClients) GetNode(localEndpoint bool) (*ethclient.Client, error) {
	addr, _, err := endpoints.Peek(localEndpoint)
	if err != nil {
		return nil, err
	}
	logger.Sugar().Infof("peek %v server", addr)
	return ethclient.Dial(addr)
}

func (eClients *EClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
	var client *ethclient.Client
	var err error
	var retry bool
	localEndpoint := true
	for i := 0; i < MaxRetries; i++ {
		client, err = eClients.GetNode(localEndpoint)
		localEndpoint = false
		if err != nil || client == nil {
			continue
		}

		retry, err = fn(ctx, client)
		if err == nil || !retry {
			return err
		}
	}
	return err
}

func (eClients EClients) BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		syncRet, _err := c.SyncProgress(ctx)
		if _err != nil {
			return true, _err
		}
		if syncRet != nil {
			return true, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.CurrentBlock, syncRet.HighestBlock,
			)
		}

		ret, err = c.BalanceAt(ctx, account, blockNumber)
		return true, err
	})

	return ret, err
}

func (eClients EClients) PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error) {
	var ret uint64
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.PendingNonceAt(ctx, account)
		return true, err
	})

	return ret, err
}

func (eClients EClients) NetworkIDS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.NetworkID(ctx)
		return true, err
	})

	return ret, err
}

func (eClients EClients) SuggestGasPriceS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.SuggestGasPrice(ctx)
		return true, err
	})

	return ret, err
}

func (eClients EClients) SendTransactionS(ctx context.Context, tx *types.Transaction) error {
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		err = c.SendTransaction(ctx, tx)
		if err != nil && (strings.Contains(err.Error(), ErrFundsToLow) || strings.Contains(err.Error(), ErrGasToLow)) {
			return false, err
		}
		return true, err
	})

	return err
}

func (eClients EClients) TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		tx, isPending, err = c.TransactionByHash(ctx, hash)
		return true, err
	})
	return tx, isPending, err
}

func (eClients EClients) TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var ret *types.Receipt
	var err error
	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.TransactionReceipt(ctx, txHash)
		return true, err
	})
	return ret, err
}

func Client() EClientI {
	return &EClients{}
}
