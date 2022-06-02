package tron

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
)

func Init() {
	rand.Seed(time.Now().Unix())
}

const (
	MinNodeNum = 1
	MaxRetries = 3
)

type ClientI interface {
	TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error)
	TRC20SendS(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetNode() *tronclient.GrpcClient
}

type TClients struct{}

func (tClients *TClients) GetNode(retries int) (*tronclient.GrpcClient, error) {
	var addr string
	var err error
	if retries == 0 {
		addr, err = endpoints.PeekPri()
	} else {
		addr, err = endpoints.Peek()
	}
	if err != nil {
		return nil, err
	}
	ntc := tronclient.NewGrpcClientWithTimeout(addr, 10*time.Second)
	err = ntc.Start(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return ntc, nil
}

func (tClients *TClients) withClient(retries int, fn func(*tronclient.GrpcClient) error) error {
	client, err := tClients.GetNode(retries)
	if err != nil {
		return err
	}
	defer client.Stop()
	return fn(client)
}

func (tClients *TClients) TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error) {
	var ret *big.Int
	var err error
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(i, func(client *tronclient.GrpcClient) error {
			ret, err = client.TRC20ContractBalance(addr, contractAddress)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20ContractBalanceS, %v", err)
}

func (tClients *TClients) TRC20SendS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(i, func(client *tronclient.GrpcClient) error {
			ret, err = client.TRC20Send(from, to, contract, amount, feeLimit)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20SendS, %v", err)
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	var ret *api.Return
	var err error
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(i, func(client *tronclient.GrpcClient) error {
			ret, err = client.Broadcast(tx)
			return err
		})
		if err == nil {
			return ret, nil
		}
		if err != nil && ret.GetCode() == api.Return_TRANSACTION_EXPIRATION_ERROR {
			return ret, err
		}
	}
	return nil, fmt.Errorf("fail BroadcastS, %v", err)
}

func (tClients *TClients) GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error) {
	var ret *core.TransactionInfo
	var err error
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(i, func(client *tronclient.GrpcClient) error {
			ret, err = client.GetTransactionInfoByID(id)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail GetTransactionInfoByIDS, %v", err)
}

var tClients *TClients

func Client() (*TClients, error) {
	if tClients != nil {
		return tClients, nil
	}

	tClients = &TClients{}
	return tClients, nil
}
