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

const (
	MinNodeNum  = 1
	MaxRetryNum = 2
)

type ClientI interface {
	TRC20ContractBalance(addr, contractAddress string) (*big.Int, error)
	TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error)
	TRC20Send(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	TRC20SendS(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	Broadcast(tx *core.Transaction) (*api.Return, error)
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByID(id string) (*core.TransactionInfo, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetNode() *tronclient.GrpcClient
}

type TClients struct {
	EndList  []string
	RetryNum uint
	ClientI
}

func NewTClients(retryNum uint, addrList []string) (*TClients, error) {
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
	rand.Seed(time.Now().Unix())
	rIndex := rand.Intn(len(tClients.EndList))
	addr := tClients.EndList[rIndex]
	ntc := tronclient.NewGrpcClientWithTimeout(addr, 10*time.Second)
	err := ntc.Start(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return ntc, nil
}

func (tClients *TClients) TRC20ContractBalance(addr, contractAddress string) (*big.Int, error) {
	client, err := tClients.GetNode()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	return client.TRC20ContractBalance(addr, contractAddress)
}

func (tClients *TClients) TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error) {
	for i := 0; i < int(tClients.RetryNum); i++ {
		ret, err := tClients.TRC20ContractBalance(addr, contractAddress)
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20ContractBalanceS")
}

func (tClients *TClients) TRC20Send(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	client, err := tClients.GetNode()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	return client.TRC20Send(from, to, contract, amount, feeLimit)
}

func (tClients *TClients) TRC20SendS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	for i := 0; i < int(tClients.RetryNum); i++ {
		ret, err := tClients.TRC20Send(from, to, contract, amount, feeLimit)
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail TRC20SendS")
}

func (tClients *TClients) Broadcast(tx *core.Transaction) (*api.Return, error) {
	client, err := tClients.GetNode()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	return client.Broadcast(tx)
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	for i := 0; i < int(tClients.RetryNum); i++ {
		ret, err := tClients.Broadcast(tx)
		if err == nil {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("fail BroadcastS")
}

func (tClients *TClients) GetTransactionInfoByID(id string) (*core.TransactionInfo, error) {
	client, err := tClients.GetNode()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	return client.GetTransactionInfoByID(id)
}

func (tClients *TClients) GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error) {
	for i := 0; i < int(tClients.RetryNum); i++ {
		ret, err := tClients.GetTransactionInfoByID(id)
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
	var err error
	tClients, err = NewTClients(MaxRetryNum, addrList)
	return tClients, err
}
