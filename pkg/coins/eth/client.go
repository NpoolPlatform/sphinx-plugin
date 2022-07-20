package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
)

var (
	gasToLow   = `intrinsic gas too low`
	fundsToLow = `insufficient funds for gas * price + value`
	nonceToLow = `nonce too low`
	stopErrMsg = []string{gasToLow, fundsToLow, nonceToLow}
)

type EClientI interface {
	BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error)
	NetworkIDS(ctx context.Context) (*big.Int, error)
	SuggestGasPriceS(ctx context.Context) (*big.Int, error)
	SendTransactionS(ctx context.Context, tx *types.Transaction) error
	TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetNode(endpointmgr *endpoints.Manager) (*ethclient.Client, error)
	WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error
}

type eClients struct{}

func (eClients eClients) GetNode(endpointmgr *endpoints.Manager) (*ethclient.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}
	return ethclient.Dial(endpoint)
}

func (eClients *eClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
	var err error
	var retry bool
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}
	for i := 0; i < MaxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}
		client, nodeErr := eClients.GetNode(endpointmgr)
		if err == nil || nodeErr != endpoints.ErrEndpointExhausted {
			err = nodeErr
		}
		if nodeErr != nil || client == nil {
			continue
		}

		retry, err = fn(ctx, client)
		client.Close()
		if err == nil || !retry {
			return err
		}
	}
	return err
}

func (eClients eClients) BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
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

func (eClients eClients) PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error) {
	var ret uint64
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.PendingNonceAt(ctx, account)
		return true, err
	})

	return ret, err
}

func (eClients eClients) NetworkIDS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.NetworkID(ctx)
		return true, err
	})

	return ret, err
}

func (eClients eClients) SuggestGasPriceS(ctx context.Context) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.SuggestGasPrice(ctx)
		return true, err
	})

	return ret, err
}

func (eClients eClients) SendTransactionS(ctx context.Context, tx *types.Transaction) error {
	var err error

	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		err = c.SendTransaction(ctx, tx)

		if err != nil && TxFailErr(err) {
			return false, err
		}
		return true, err
	})

	return err
}

func (eClients eClients) TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		tx, isPending, err = c.TransactionByHash(ctx, hash)
		return true, err
	})
	return tx, isPending, err
}

func (eClients eClients) TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var ret *types.Receipt
	var err error
	err = eClients.WithClient(ctx, func(ctx context.Context, c *ethclient.Client) (bool, error) {
		ret, err = c.TransactionReceipt(ctx, txHash)
		return true, err
	})
	return ret, err
}

func Client() EClientI {
	return &eClients{}
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
