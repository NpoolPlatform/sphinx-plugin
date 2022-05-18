package tron

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
)

const MIN_NODE_NUM = 1

type TronClientI interface {
	TRC20ContractBalance(addr, contractAddress string) (*big.Int, error)
	TRC20Send(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
	Broadcast(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByID(id string) (*core.TransactionInfo, error)
	OnFailed(*tronclient.GrpcClient)
}

type TClient struct {
	GrpcClient *tronclient.GrpcClient
	FailedNum  int
}

type TronClient struct {
	TClientList []*TClient
	RetryNum    uint
	TronClientI
}

func NewTronClient(retryNum uint, addrList []string) (*TronClient, error) {
	tronClient := &TronClient{}
	tronClient.RetryNum = retryNum
	for _, addr := range addrList {
		client := &TClient{}
		client.GrpcClient = tronclient.NewGrpcClient(addr)
		err := client.GrpcClient.Start(grpc.WithInsecure())
		if err != nil {
			continue
		}
		tronClient.TClientList = append(tronClient.TClientList, client)
	}
	if len(tronClient.TClientList) < MIN_NODE_NUM {
		return tronClient, fmt.Errorf("too few nodes have been successfully connected,just %v nodes",
			len(tronClient.TClientList))
	}
	return tronClient, nil
}

var tronClient *tronclient.GrpcClient

// TODO main init env and check, use conn pool
func client() (*tronclient.GrpcClient, error) {
	// TODO all env use cache
	endpoint, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	client := tronclient.NewGrpcClient(endpoint)
	err := client.Start(grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc client start error: %v", err)
	}

	return client, nil
}

func keepConnect(tronClient *tronclient.GrpcClient) error {
	_, err := tronClient.GetNodeInfo()
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return tronClient.Reconnect(tronClient.Address)
		}
		return fmt.Errorf("node connect error: %v", err)
	}
	return nil
}

func Client() (*tronclient.GrpcClient, error) {
	if tronClient != nil {
		err := keepConnect(tronClient)
		if err == nil {
			return tronClient, nil
		}
		err = tronClient.Conn.Close()
		if err != nil {
			return nil, err
		}
	}
	tronClient, err := client()
	if err != nil {
		return nil, err
	}
	return tronClient, nil
}
