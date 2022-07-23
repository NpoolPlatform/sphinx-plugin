package sol

import (
	"context"
	"errors"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
	"github.com/gagliardetto/solana-go/rpc"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
)

type SClientI interface {
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*rpc.Client, error)
	WithClient(ctx context.Context, fn func(*rpc.Client) (bool, error)) error
}

type SClients struct{}

func (sClients SClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*rpc.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}

	client := rpc.New(endpoint)
	_, err = client.GetHealth(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (sClients *SClients) WithClient(ctx context.Context, fn func(c *rpc.Client) (bool, error)) error {
	var (
		apiErr, err error
		retry       bool
		client      *rpc.Client
	)
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}

	for i := 0; i < utils.MinInt(MaxRetries, endpointmgr.Len()); i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		client, err = sClients.GetNode(ctx, endpointmgr)
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

		if !retry {
			return apiErr
		}
	}
	return err
}

func Client() SClientI {
	return &SClients{}
}
