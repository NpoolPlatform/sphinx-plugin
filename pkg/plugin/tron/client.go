package tron

import (
	"fmt"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
)

var tronClient *tronclient.GrpcClient

// TODO main init env and check, use conn pool
func client() (*tronclient.GrpcClient, error) {
	// TODO all env use cache
	endpoint, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	client := tronclient.NewGrpcClient(endpoint)
	err := client.Start(grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc client start error: %v", err)
	}

	return client, nil
}

func Client() (*tronclient.GrpcClient, error) {
	if tronClient != nil {
		return tronClient, nil
	}
	var err error
	tronClient, err = client()
	if err != nil {
		return nil, err
	}
	return tronClient, nil
}
