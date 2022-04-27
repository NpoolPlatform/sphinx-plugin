package tron

import (
	"fmt"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	tronclent "github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
)

var tronClient *tronclent.GrpcClient

// TODO main init env and check, use conn pool
func client() (*tronclent.GrpcClient, error) {
	// TODO all env use cache
	endpoint, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	client := tronclent.NewGrpcClient(endpoint)
	err := client.Start(grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc client start error: %v", err)
	}

	return client, nil
}

func keepConnect(tronClient *tronclent.GrpcClient) error {
	_, err := tronClient.GetNodeInfo()
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return tronClient.Reconnect(tronClient.Address)
		}
		return fmt.Errorf("node connect error: %v", err)
	}
	return nil
}

func Client() (*tronclent.GrpcClient, error) {
	if tronClient != nil {
		return tronClient, keepConnect(tronClient)
	}
	return client()
}
