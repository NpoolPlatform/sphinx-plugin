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

func init() {
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

type TClients struct {
	localEndpoint bool
}

func (tClients *TClients) GetNode() (*tronclient.GrpcClient, error) {
	addr, err := endpoints.Peek(tClients.localEndpoint)
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

func (tClients *TClients) withClient(fn func(*tronclient.GrpcClient) (bool, error)) error {
	localEndpoint := true
	for i := 0; i < MaxRetries; i++ {
		client, err := tClients.Peek(localEndpoint)
		if err != nil {
			return err
		}
		retry, err := fn(client)
		if err != nil {
			if !retry {
				client.Stop()
				return fmt.Errorf("fail run action: %v", err)
			}
		}
		client.Stop()
		localEndpoint = false
	}
}

func (tClients *TClients) TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error) {
	var ret *big.Int
	var err error

	err = tClients.withClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.TRC20ContractBalance(addr, contractAddress)
		return false, err
	})
	if err == nil {
		return ret, nil
	}
	
	return nil, fmt.Errorf("fail TRC20ContractBalanceS, %v", err)
}

func (tClients *TClients) TRC20SendS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error

	tClients.localEndpoint = true
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.TRC20Send(from, to, contract, amount, feeLimit)
			return err
		})
		if err == nil {
			return ret, nil
		}
		tClients.localEndpoint = false
	}
	return nil, fmt.Errorf("fail TRC20SendS, %v", err)
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	var ret *api.Return
	var err error

	tClients.localEndpoint = true
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.Broadcast(tx)
			return err
		})
		if err == nil {
			return ret, nil
		}
		if err != nil && ret.GetCode() == api.Return_TRANSACTION_EXPIRATION_ERROR {
			return ret, err
		}
		tClients.localEndpoint = false
	}
	return nil, fmt.Errorf("fail BroadcastS, %v", err)
}

func (tClients *TClients) GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error) {
	var ret *core.TransactionInfo
	var err error

	tClients.localEndpoint = true
	for i := 0; i < MaxRetries; i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.GetTransactionInfoByID(id)
			return err
		})
		if err == nil {
			return ret, nil
		}
		tClients.localEndpoint = false
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
