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
	ErrGasToLow   = `intrinsic gas too low`
	ErrFundsToLow = `insufficient funds for gas * price + value`
	ErrNonceToLow = `nonce too low`
	StopErrs      = []string{ErrGasToLow, ErrFundsToLow, ErrNonceToLow}
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

type EClients struct{}

func (eClients EClients) GetNode(endpointmgr *endpoints.Manager) (*ethclient.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}
	return ethclient.Dial(endpoint)
}

func (eClients *EClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
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
		defer client.Close()

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

		if err != nil && TxFailErr(err) {
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

func TxFailErr(err error) bool {
	for _, v := range StopErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
