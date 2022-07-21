package fil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
	EndpointSep      = `|`
	EndpointInvalid  = `fil endpoint invalid`
)

type FClientI interface {
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (v0api.FullNode, jsonrpc.ClientCloser, error)
	WithClient(ctx context.Context, fn func(v0api.FullNode) (bool, error)) error
}

type FClients struct{}

func (fClients FClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (v0api.FullNode, jsonrpc.ClientCloser, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, nil, err
	}

	strs := strings.Split(endpoint, EndpointSep)
	if len(strs) != 2 {
		return nil, nil, fmt.Errorf("%v,%v", EndpointInvalid, endpoint)
	}

	addr := strs[0]
	authToken := strs[1]
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	var _api v0api.FullNodeStruct
	// internal has conn pool
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", lotusapi.GetInternalStructs(&_api), headers)
	if err != nil {
		return nil, nil, err
	}

	return &_api, closer, nil
}

func (fClients *FClients) WithClient(ctx context.Context, fn func(c v0api.FullNode) (bool, error)) error {
	var (
		apiErr, err error
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

		client, closer, err := fClients.GetNode(ctx, endpointmgr)
		if errors.Is(err, endpoints.ErrEndpointExhausted) {
			return apiErr
		}

		if err != nil {
			closer()
			return err
		}

		retry, apiErr = fn(client)
		closer()

		if err == nil || !retry {
			return err
		}
	}
	return err
}

func Client() FClientI {
	return &FClients{}
}