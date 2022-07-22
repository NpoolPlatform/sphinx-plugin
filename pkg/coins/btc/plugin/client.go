package plugin

// import (
// 	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
// 	"github.com/btcsuite/btcd/chaincfg"
// )

// var configParam = map[string]string{
// 	coins.CoinNetMain: chaincfg.MainNetParams.Name,
// 	coins.CoinNetTest: chaincfg.RegressionNetParams.Name,
// }

// // TODO main init env and check, use conn pool
// func client() (*rpcclient.Client, error) {
// 	/*
// 		Host:         "localhost:8332",
// 		User:         "yourrpcuser",
// 		Pass:         "yourrpcpass",
// 		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
// 		DisableTLS:   true, // Bitcoin core does not provide TLS by default
// 	*/

// 	// TODO all env use cache
// 	host, ok := env.LookupEnv(env.ENVCOINLOCALAPI)
// 	if !ok {
// 		return nil, env.ErrENVCoinLocalAPINotFound
// 	}
// 	user, ok := env.LookupEnv(env.ENVCOINUSER)
// 	if !ok {
// 		return nil, env.ErrENVCoinUserNotFound
// 	}
// 	pass, ok := env.LookupEnv(env.ENVCOINPASS)
// 	if !ok {
// 		return nil, env.ErrENVCoinPassNotFound
// 	}

// 	v, ok := env.LookupEnv(env.ENVCOINNET)
// 	if !ok {
// 		return nil, env.ErrEVNCoinNet
// 	}
// 	if !coins.CheckSupportNet(v) {
// 		return nil, env.ErrEVNCoinNetValue
// 	}

// 	connCfg := &rpcclient.ConnConfig{
// 		Host: host,
// 		User: user,
// 		Pass: pass,
// 		// default mainnet
// 		Params:       configParam[v],
// 		HTTPPostMode: true,
// 		DisableTLS:   true,
// 	}

// 	cli, err := rpcclient.New(connCfg, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return cli, nil
// }
