package eth

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO main init env and check, use conn pool
func client() (*ethclient.Client, error) {
	// TODO all env use cache
	endpoint, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}

	return ethclient.Dial(endpoint)
}

func Client() (*ethclient.Client, error) {
	return client()
}

// func Init() {
// 	rand.Seed(time.Now().Unix())
// }

// const (
// 	MinNodeNum  = 1
// 	MaxRetryNum = 3
// )

// var (
// 	ErrGasToLow   = "intrinsic gas too low"
// 	ErrFundsToLow = "insufficient funds for gas * price + value"
// )

// type BClientI interface {
// 	BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
// 	PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error)
// 	NetworkIDS(ctx context.Context) (*big.Int, error)
// 	SuggestGasPriceS(ctx context.Context) (*big.Int, error)
// 	SendTransactionS(ctx context.Context, tx *types.Transaction) error
// 	TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
// 	TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
// 	GetNode() (*ethclient.Client, error)
// }

// type BClients struct {
// 	EndPoints []string
// 	RetryNum  int
// }

// func (bClients BClients) GetNode() (*ethclient.Client, error) {
// 	randI := rand.Intn(len(bClients.EndPoints))
// 	addr := bClients.EndPoints[randI]
// 	return ethclient.Dial(addr)
// }

// func (bClients *BClients) withClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) error) error {
// 	client, err := bClients.GetNode()
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()

// 	return fn(ctx, client)
// }

// func (bClients BClients) BalanceAtS(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
// 	var ret *big.Int
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			ret, err = c.BalanceAt(ctx, account, blockNumber)
// 			return err
// 		})
// 		if err == nil {
// 			return ret, nil
// 		}
// 	}
// 	return ret, fmt.Errorf("fail BlanceAtS, %v", err)
// }

// func (bClients BClients) PendingNonceAtS(ctx context.Context, account common.Address) (uint64, error) {
// 	var ret uint64
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			ret, err = c.PendingNonceAt(ctx, account)
// 			return err
// 		})
// 		if err == nil {
// 			return ret, nil
// 		}
// 	}
// 	return ret, fmt.Errorf("fail PendingNonceAtS, %v", err)
// }

// func (bClients BClients) NetworkIDS(ctx context.Context) (*big.Int, error) {
// 	var ret *big.Int
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			ret, err = c.NetworkID(ctx)
// 			return err
// 		})
// 		if err == nil {
// 			return ret, nil
// 		}
// 	}
// 	return ret, fmt.Errorf("fail NetworkIDS, %v", err)
// }

// func (bClients BClients) SuggestGasPriceS(ctx context.Context) (*big.Int, error) {
// 	var ret *big.Int
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			ret, err = c.SuggestGasPrice(ctx)
// 			return err
// 		})
// 		if err == nil {
// 			return ret, nil
// 		}
// 	}
// 	return ret, fmt.Errorf("fail SuggestGasPriceS, %v", err)
// }

// func (bClients BClients) SendTransactionS(ctx context.Context, tx *types.Transaction) error {
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			err = c.SendTransaction(ctx, tx)
// 			return err
// 		})
// 		if err == nil {
// 			return nil
// 		}
// 		if strings.Contains(err.Error(), ErrFundsToLow) || strings.Contains(err.Error(), ErrGasToLow) {
// 			break
// 		}
// 	}
// 	return fmt.Errorf("fail SendTransactionS, %v", err)
// }

// func (bClients BClients) TransactionByHashS(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			tx, isPending, err = c.TransactionByHash(ctx, hash)
// 			return err
// 		})
// 		if err == nil {
// 			return tx, isPending, nil
// 		}
// 	}
// 	return tx, isPending, fmt.Errorf("fail TransactionByHashS, %v", err)
// }

// func (bClients BClients) TransactionReceiptS(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
// 	var ret *types.Receipt
// 	var err error
// 	for i := 0; i < bClients.RetryNum; i++ {
// 		err = bClients.withClient(ctx, func(ctx context.Context, c *ethclient.Client) error {
// 			ret, err = c.TransactionReceipt(ctx, txHash)
// 			return err
// 		})
// 		if err == nil {
// 			return ret, nil
// 		}
// 	}
// 	return ret, fmt.Errorf("fail TransactionReceiptS, %v", err)
// }

// func newBSCClients(retryNum int, endpoints []string) (*BClients, error) {
// 	bscClients := &BClients{}
// 	bscClients.RetryNum = retryNum
// 	for _, endpoint := range endpoints {
// 		client, err := ethclient.Dial(endpoint)
// 		if err != nil {
// 			continue
// 		}
// 		client.Close()
// 		bscClients.EndPoints = append(bscClients.EndPoints, endpoint)
// 	}
// 	if len(bscClients.EndPoints) < MinNodeNum {
// 		return bscClients, fmt.Errorf("too few nodes have been successfully connected,just %v nodes",
// 			len(bscClients.EndPoints))
// 	}
// 	return bscClients, nil
// }

// var bscClients *BClients

// func Client() (*BClients, error) {
// 	if bscClients != nil {
// 		return bscClients, nil
// 	}
// 	addrs, ok := env.LookupEnv(env.ENVCOINAPI)
// 	if !ok {
// 		return nil, env.ErrENVCoinAPINotFound
// 	}
// 	endpoints := strings.Split(addrs, ",")
// 	var err error
// 	bscClients, err = newBSCClients(MaxRetryNum, endpoints)
// 	return bscClients, err
// }
