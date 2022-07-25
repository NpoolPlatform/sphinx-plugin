package bsc

import (
	"context"
	"fmt"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/utils"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MinNodeNum = 1
	MaxRetries = 3

	retriesSleepTime = 200 * time.Millisecond
	dialTimeout      = 3 * time.Second
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

	ctx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

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
		client      *ethclient.Client
	)
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}
	for i := 0; i < utils.MinInt(MaxRetries, endpointmgr.Len()); i++ {
		if i > 0 {
			time.Sleep(retriesSleepTime)
		}

		client, err = bClients.GetNode(ctx, endpointmgr)
		if err != nil {
			continue
		}

		retry, apiErr = fn(ctx, client)
		client.Close()

		if !retry {
			return apiErr
		}
	}
	if apiErr != nil {
		return apiErr
	}
	return err
}

func Client() BClientI {
	return &bClients{}
}
