package bsc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MinNodeNum = 1
	MaxRetries = 3
)

type BClientI interface {
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*ethclient.Client, error)
	WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error
}

type bClients struct{}

func (bClients bClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*ethclient.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}
	cli, err := ethclient.DialContext(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	syncRet, _err := cli.SyncProgress(ctx)
	if _err != nil {
		cli.Close()
		return nil, _err
	}

	if syncRet != nil {
		cli.Close()
		return nil, fmt.Errorf(
			"node is syncing ,current block %v ,highest block %v ",
			syncRet.CurrentBlock, syncRet.HighestBlock,
		)
	}

	return cli, nil
}

func (bClients *bClients) WithClient(ctx context.Context, fn func(ctx context.Context, c *ethclient.Client) (bool, error)) error {
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

		client, err := bClients.GetNode(ctx, endpointmgr)

		if errors.Is(err, endpoints.ErrEndpointExhausted) {
			if apiErr != nil {
				return apiErr
			}
			return err
		}
		if err != nil {
			continue
		}

		retry, apiErr = fn(ctx, client)
		client.Close()

		if !retry {
			return apiErr
		}
	}
	return err
}

func Client() BClientI {
	return &bClients{}
}
