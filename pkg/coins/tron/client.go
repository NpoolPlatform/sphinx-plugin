package tron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
)

type TClientI interface {
	GetGRPCClient(timeout time.Duration, endpointmgr *endpoints.Manager) (*tronclient.GrpcClient, error)
	WithClient(fn func(*tronclient.GrpcClient) (bool, error)) error
}

type tClients struct{}

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

func (tClients *tClients) GetGRPCClient(timeout time.Duration, endpointmgr *endpoints.Manager) (*tronclient.GrpcClient, error) {
	endpoint, isLocal, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}

	if isLocal {
		strs := strings.Split(endpoint, ":")
		port := jsonAPIMap[strs[0]]

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		syncRet, _err := tClients.SyncProgress(ctx, strs[0], port)
		if _err != nil {
			return nil, _err
		}

		if syncRet != nil {
			return nil, fmt.Errorf(
				"node is syncing ,current block %v ,highest block %v ",
				syncRet.Result.CurrentBlock, syncRet.Result.HighestBlock,
			)
		}
	}

	ntc := tronclient.NewGrpcClientWithTimeout(endpoint, timeout)
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

var client = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (tClients *tClients) SyncProgress(ctx context.Context, ip, port string) (*SyncingResponse, error) {
	addr := fmt.Sprintf("http://%v:%v/jsonrpc", ip, port)
	req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":64}`))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	syncRes := SyncingResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&syncRes); err != nil {
		return nil, err
	}

	if syncRes.Result.CurrentBlock >= syncRes.Result.HighestBlock {
		return nil, nil
	}

	return &syncRes, nil
}

func (tClients *tClients) WithClient(fn func(*tronclient.GrpcClient) (bool, error)) error {
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
		client, err := tClients.GetGRPCClient(6*time.Second, endpointmgr)
		if errors.Is(err, endpoints.ErrEndpointExhausted) {
			if apiErr != nil {
				return apiErr
			}
			return err
		}

		if err != nil {
			continue
		}

		retry, apiErr = fn(client)
		client.Stop()

		if apiErr == nil || !retry {
			return apiErr
		}
	}

	return err
}

func Client() TClientI {
	return &tClients{}
}
