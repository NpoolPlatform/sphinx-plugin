package tron

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
)

const (
	TxExpired  = `Transaction expired`
	FundsToLow = `balance is not sufficient`
)

var StopErrs = []string{TxExpired, FundsToLow}

type TClientI interface {
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
	var (
		err, apiErr error
		retry       bool
	)

	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}
	for i := 0; i < MaxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}
		client, err := tClients.GetGRPCClient(endpointmgr)
		if errors.Is(err, endpoints.ErrEndpointExhausted) {
			return apiErr
		}

		if err != nil {
			return err
		}
		retry, err = fn(client)
		client.Stop()

		if err == nil || !retry {
			return err
		}
	}
	return err
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
