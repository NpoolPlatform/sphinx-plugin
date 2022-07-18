package tron

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
)

var (
	ErrTxExpired  = `Transaction expired`
	ErrFundsToLow = `balance is not sufficient`
	StopErrs      = []string{ErrTxExpired, ErrFundsToLow}
)

type TClientI interface {
	TRXBalanceS(addr string) (int64, error)
	TRXTransferS(from, to string, amount int64) (*api.TransactionExtention, error)
	BroadcastS(tx *core.Transaction) (*api.Return, error)
	GetTransactionInfoByIDS(id string) (*core.TransactionInfo, error)
	GetGRPCClient(endpointmgr *endpoints.Manager) (*tronclient.GrpcClient, error)
	WithClient(fn func(*tronclient.GrpcClient) (bool, error)) error
}

type TClients struct{}

var jsonAPIMap map[string]string

func init() {
	jsonAPIMap = make(map[string]string)
	var jsonApis []string

	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCLOCALAPI); ok {
		strs := strings.Split(v, endpoints.AddrSplitter)
		jsonApis = append(jsonApis, strs...)
	}
	if v, ok := env.LookupEnv(env.ENVCOINJSONRPCPUBLICAPI); ok {
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

func (tClients *TClients) GetGRPCClient(endpointmgr *endpoints.Manager) (*tronclient.GrpcClient, error) {
	endpoint, isLocal, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}

	if isLocal {
		strs := strings.Split(endpoint, ":")
		port := jsonAPIMap[strs[0]]
		syncRet, _err := tClients.SyncProgress(strs[0], port)
		if _err != nil {
			log.Error(_err)
			return nil, _err
		}

		if syncRet != nil {
			return nil, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.Result.CurrentBlock, syncRet.Result.HighestBlock,
			)
		}
	}

	ntc := tronclient.NewGrpcClientWithTimeout(endpoint, 6*time.Second)
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
		client, nodeErr := tClients.GetGRPCClient(endpointmgr)
		if err == nil || nodeErr != endpoints.ErrEndpointExhausted {
			err = nodeErr
		}
		if nodeErr != nil || client == nil {
			continue
		}
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
	ret := EmptyTRX
	if err := ValidAddress(addr); err != nil {
		return ret, err
	}

	err := tClients.WithClient(func(client *tronclient.GrpcClient) (bool, error) {
		acc, err := client.GetAccount(addr)
		if err != nil {
			return true, err
		}
		ret = acc.GetBalance()
		return false, nil
	})
	if err != nil && strings.Contains(err.Error(), ErrInvalidAddr.Error()) {
		return EmptyTRX, nil
	}
	return ret, err
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

func TxFailErr(err error) bool {
	for _, v := range StopErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
