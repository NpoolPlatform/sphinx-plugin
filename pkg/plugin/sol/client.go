package sol

import (
	"context"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/gagliardetto/solana-go/rpc"
)

// TODO:Now is a compromise, consider the way of using the pool of clients.
var rpcClent *rpc.Client

func NewClient() (*rpc.Client, error) {
	addr, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	client := rpc.New(addr)

	client.GetHealth(context.Background())
	return client, nil
}

func client() (*rpc.Client, error) {
	if rpcClent == nil {
		rpcClient, err := NewClient()
		return rpcClient, err
	}
	return rpcClent, nil
}
