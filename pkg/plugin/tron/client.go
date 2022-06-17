package tron

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
<<<<<<< HEAD
	TRC20ContractBalanceS(addr, contractAddress string) (*big.Int, error)
	TRC20SendS(from string, to string, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error)
=======
	TRXBalanceS(addr string) (int64, error)
	TRXTransferS(from, to string, amount int64) (*api.TransactionExtention, error)
>>>>>>> current err of nil and support check tron-account
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetGRPCClient(localEndpoint bool) (*tronclient.GrpcClient, error)
	WithClient(fn func(*tronclient.GrpcClient) (bool, error)) error
}

type TClients struct{}

var jsonAPIMap map[string]string

func init() {
	jsonAPIMap = make(map[string]string)
	var jsonApis []string
<<<<<<< HEAD

	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCLOCALPORT); ok {
		strs := strings.Split(v, endpoints.AddrSplitter)
		jsonApis = append(jsonApis, strs...)
	}

	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCPUBLICPORT); ok {
=======
	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCLOCALAPI); ok {
		strs := strings.Split(v, endpoints.AddrSplitter)
		jsonApis = append(jsonApis, strs...)
	}
	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCPUBLICAPI); ok {
>>>>>>> current err of nil and support check tron-account
		strs := strings.Split(v, endpoints.AddrSplitter)
		jsonApis = append(jsonApis, strs...)
	}

	for _, v := range jsonApis {
		strs := strings.Split(v, ":")
		if len(strs) < 2 {
			continue
		}
		jsonAPIMap[strs[0]] = strs[1]
	}
}

func (tClients *TClients) GetGRPCClient(localEndpoint bool) (*tronclient.GrpcClient, error) {
	addr, isLocal, err := endpoints.Peek(localEndpoint)
	if err != nil {
		return nil, err
	}
	strs := strings.Split(addr, ":")

	if isLocal {
		port := jsonAPIMap[strs[0]]
		syncRet, _err := tClients.SyncProgress(strs[0], port)
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
	}

	ntc := tronclient.NewGrpcClientWithTimeout(addr, 6*time.Second)
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

func (tClients *TClients) WithClient(fn func(*tronclient.GrpcClient) (bool, error)) error {
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
		ret = acc.GetBalance()
		return false, nil
	})
	if err != nil && strings.Contains(err.Error(), ErrAccountNotFound.Error()) {
		return EmptyTRX, nil
	}
	return ret, nil
}

func (tClients *TClients) TRXTransferS(from, to string, amount int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	err = tClients.WithClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.Transfer(from, to, amount)
		return true, err
	})

	return ret, err
}

func (tClients *TClients) BroadcastS(tx *core.Transaction) (*api.Return, error) {
	var ret *api.Return
	var err error

	err = tClients.WithClient(func(client *tronclient.GrpcClient) (bool, error) {
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

	err = tClients.WithClient(func(client *tronclient.GrpcClient) (bool, error) {
		ret, err = client.GetTransactionInfoByID(id)
		return true, err
	})

	return ret, err
}

func Client() TClientI {
	return &TClients{}
}
