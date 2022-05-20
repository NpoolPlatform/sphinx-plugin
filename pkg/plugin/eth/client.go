package eth

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO main init env and check, use conn pool
func client() (*ethclient.Client, error) {
	// TODO all env use cache
	endpoint, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}

	return ethclient.Dial(endpoint)
}

func Client() (*ethclient.Client, error) {
	return client()
}
