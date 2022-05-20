package tron

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
)

func Init() {
	rand.Seed(time.Now().Unix())
}

const (
	MinNodeNum  = 1
	MaxRetryNum = 2
)

type ClientI interface {
	TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error)
	TRC20SendS(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetNode() *tronclient.GrpcClient
}

type TClients struct {
	EndList  []string
	RetryNum uint
}

func newTClients(retryNum uint, addrList []string) (*TClients, error) {
	tronClient := &TClients{}
	tronClient.RetryNum = retryNum
	for _, addr := range addrList {
		client := tronclient.NewGrpcClient(addr)
		err := client.Start(grpc.WithInsecure())
		if err != nil {
			continue
		}
		client.Stop()
		tronClient.EndList = append(tronClient.EndList, addr)
	}
	if len(tronClient.EndList) < MinNodeNum {
		return tronClient, fmt.Errorf("too few nodes have been successfully connected,just %v nodes",
			len(tronClient.EndList))
	}
	return tronClient, nil
}

func (tClients *TClients) GetNode() (*tronclient.GrpcClient, error) {
	rIndex := rand.Intn(len(tClients.EndList))
	addr := tClients.EndList[rIndex]
	ntc := tronclient.NewGrpcClientWithTimeout(addr, 10*time.Second)
	err := ntc.Start(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return ntc, nil
}

func (tClients *TClients) withClient(fn func(*tronclient.GrpcClient) error) error {
	client, err := tClients.GetNode()
	if err != nil {
		return err
	}
	defer client.Stop()
	return fn(client)
}

func (tClients *TClients) TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error) {
	var ret *big.Int
	var err error
	for i := 0; i < int(tClients.RetryNum); i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.TRC20ContractBalance(addr, contractAddress)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20ContractBalanceS")
}

func (tClients *TClients) TRC20SendS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	for i := 0; i < int(tClients.RetryNum); i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.TRC20Send(from, to, contract, amount, feeLimit)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20SendS")
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	var ret *api.Return
	var err error
	for i := 0; i < int(tClients.RetryNum); i++ {
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
	}
	return nil, fmt.Errorf("fail BroadcastS")
}

func (tClients *TClients) GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error) {
	var ret *core.TransactionInfo
	var err error
	for i := 0; i < int(tClients.RetryNum); i++ {
		err = tClients.withClient(func(client *tronclient.GrpcClient) error {
			ret, err = client.GetTransactionInfoByID(id)
			return err
		})
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail GetTransactionInfoByIDS")
}

var tClients *TClients

func Client() (*TClients, error) {
	if tClients != nil {
		return tClients, nil
	}
	addrs, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	addrList := strings.Split(addrs, ",")
	return newTClients(MaxRetryNum, addrList)
}
