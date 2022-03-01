package btc

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

var configParam = map[string]string{
	plugin.CoinNetMain: chaincfg.MainNetParams.Name,
	plugin.CoinNetTest: chaincfg.RegressionNetParams.Name,
}

// TODO main init env and check, use conn pool
func client() (*rpcclient.Client, error) {
	/*
		Host:         "localhost:8332",
		User:         "yourrpcuser",
		Pass:         "yourrpcpass",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	*/

	// TODO all env use cache
	host, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return nil, env.ErrENVCoinAPINotFound
	}
	user, ok := env.LookupEnv(env.ENVCOINUSER)
	if !ok {
		return nil, env.ErrENVCoinUserNotFound
	}
	pass, ok := env.LookupEnv(env.ENVCOINPASS)
	if !ok {
		return nil, env.ErrENVCoinPassNotFound
	}
	connCfg := &rpcclient.ConnConfig{
		Host: host,
		User: user,
		Pass: pass,
		// default mainnet
		Params:       configParam[plugin.CoinNet],
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	return rpcclient.New(connCfg, nil)
}
