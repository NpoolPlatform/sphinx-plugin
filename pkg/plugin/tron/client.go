package tron

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
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
	GetGRPCClient(localEndpoint bool) (*tronclient.GrpcClient, error)
}

type TClients struct{}

func (tClients *TClients) GetGRPCClient(localEndpoint bool) (*tronclient.GrpcClient, error) {
	grpcport, ok := env.LookupEnv(env.ENVCOINGRPCPORT)
	if !ok {
		return nil, env.ErrENVCOINGRPCPortFound
	}

	jsonrpcport, ok := env.LookupEnv(env.ENVCOINJSONRPCPORT)
	if !ok {
		return nil, env.ErrENVCOINJSONRPCPortFound
	}

	addr, err := endpoints.Peek(localEndpoint)
	if err != nil {
		return nil, err
	}
	syncRet, _err := tClients.SyncProgress(addr, jsonrpcport)

	if _err != nil {
		logger.Sugar().Error(_err)
		return nil, _err
	}
	if syncRet != nil {
		return nil, fmt.Errorf(
			"node is syncing ,current block %v ,highest block %v ",
			syncRet.Result.CurrentBlock, syncRet.Result.HighestBlock,
		)
	}

	endpoint := fmt.Sprintf("%v:%v", addr, grpcport)
	logger.Sugar().Infof("peek %v server", endpoint)

	ntc := tronclient.NewGrpcClientWithTimeout(endpoint, 10*time.Second)
	err = ntc.Start(grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return ntc, nil
}

type SyncBlock struct {
	StartingBlock string
	CurrentBlock  string
	HighestBlock  string
}

type SyncingResponse struct {
	ID     int
	Result SyncBlock
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (tClients *TClients) SyncProgress(ip, port string) (*SyncingResponse, error) {
	addr := fmt.Sprintf("http://%v:%v/jsonrpc", ip, port)
	contentType := "application/json"
	body := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":64}`)

	resp, err := http.Post(addr, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var syncRes SyncingResponse
	err = json.Unmarshal(data, &syncRes)
	if err != nil {
		return nil, err
	}
	if syncRes.Result.CurrentBlock >= syncRes.Result.HighestBlock {
		return nil, nil
	}

	return &syncRes, nil
}

func (tClients *TClients) withClient(fn func(*tronclient.GrpcClient) (bool, error)) error {
	localEndpoint := true

	var err error
	var retry bool
	var client *tronclient.GrpcClient

	for i := 0; i < MaxRetries; i++ {
		client, err = tClients.GetGRPCClient(localEndpoint)
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
