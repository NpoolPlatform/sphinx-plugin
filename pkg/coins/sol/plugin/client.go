package plugin

import (
	"context"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/gagliardetto/solana-go/rpc"
)

// TODO:Now is a compromise, consider the way of using the pool of clients.
var rpcClient *rpc.Client

func NewClient(ctx context.Context) (*rpc.Client, error) {
	addr, ok := env.LookupEnv(env.ENVCOINLOCALAPI)
	if !ok {
		return nil, env.ErrENVCoinLocalAPINotFound
	}

	client := rpc.New(addr)
	_, err := client.GetHealth(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func client(ctx context.Context) (*rpc.Client, error) {
	var err error

	if rpcClient == nil {
		rpcClient, err = NewClient(ctx)
		return rpcClient, err
	}

	_, err = rpcClient.GetHealth(ctx)
	if err != nil {
		rpcClient, err = NewClient(ctx)
		return rpcClient, err
	}

	return rpcClient, nil
}
