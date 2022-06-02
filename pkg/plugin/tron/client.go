package tron

import (
	"math/big"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
)

const (
	MinNodeNum = 1
	MaxRetries = 3
)

type TClientI interface {
	TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error)
	TRC20SendS(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetNode(localEndpoint bool) (*tronclient.GrpcClient, error)
}

type TClients struct{}

func (tClients *TClients) GetNode(localEndpoint bool) (*tronclient.GrpcClient, error) {
	addr, err := endpoints.Peek(localEndpoint)
	if err != nil {
		return nil, err
	}
	logger.Sugar().Infof("peek %v server", addr)

	ntc := tronclient.NewGrpcClientWithTimeout(addr, 10*time.Second)
	err = ntc.Start(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return ntc, nil
}

func (tClients *TClients) withClient(fn func(*tronclient.GrpcClient) (bool, error)) error {
	localEndpoint := true

	var err error
	var retry bool
	var client *tronclient.GrpcClient

	for i := 0; i < MaxRetries; i++ {
		client, err = tClients.GetNode(localEndpoint)
		localEndpoint = false
		if err != nil {
			continue
		}
		retry, err = fn(client)
		client.Stop()

		if err == nil || !retry {
			return err
		}
	}
	return err
}

func (tClients *TClients) TRXBalanceS(addr string) (int64, error) {
	var ret int64
	var err error
	for i := 0; i < int(tClients.Retries); i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			acc, err := client.GetAccount(addr)
			if err != nil {
				return err
			}
			ret = acc.GetBalance()
			return nil
		})
		if err == nil {
			return ret, nil
		}
	}
	return ret, fmt.Errorf("fail TRXBalanceS, %v", err)
}

func (tClients *TClients) TRXTransferS(from, to string, amount int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	for i := 0; i < int(tClients.Retries); i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.Transfer(from, to, amount)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRXTransferS, %v", err)
}

func (tClients *TClients) TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = tClients.withClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.TRC20ContractBalance(addr, contractAddress)
		return true, err
	})

	return ret, err
}

func (tClients *TClients) TRC20SendS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error

	err = tClients.withClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.TRC20Send(from, to, contract, amount, feeLimit)
		return true, err
	})

	return ret, err
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	var ret *api.Return
	var err error

	err = tClients.withClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.Broadcast(tx)
		if err != nil && ret.GetCode() == api.Return_TRANSACTION_EXPIRATION_ERROR {
			return false, err
		}
		return true, err
	})

	return ret, err
}

func (tClients *TClients) GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error) {
	var ret *core.TransactionInfo
	var err error

	err = tClients.withClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.GetTransactionInfoByID(id)
		return true, err
	})

	return ret, err
}

func Client() TClientI {
	return &TClients{}
}
