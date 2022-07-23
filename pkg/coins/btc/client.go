package btc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/endpoints"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
)

const (
	MinNodeNum       = 1
	MaxRetries       = 3
	RetriesSleepTime = 1 * time.Second
	EndpointSep      = `|`
	EndpointAuthSep  = `@`
	EndpointInvalid  = `fil endpoint invalid`
)

var configParam = map[string]string{
	coins.CoinNetMain: chaincfg.MainNetParams.Name,
	coins.CoinNetTest: chaincfg.RegressionNetParams.Name,
}

type BClientI interface {
	GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*rpcclient.Client, error)
	WithClient(ctx context.Context, fn func(*rpcclient.Client) (bool, error)) error
}

type BClients struct{}

func (bClients BClients) GetNode(ctx context.Context, endpointmgr *endpoints.Manager) (*rpcclient.Client, error) {
	endpoint, _, err := endpointmgr.Peek()
	if err != nil {
		return nil, err
	}
	strs := strings.Split(endpoint, EndpointSep)
	if len(strs) != 2 {
		return nil, fmt.Errorf("%v,%v", EndpointInvalid, endpoint)
	}
	host := strs[0]

	strs = strings.Split(strs[1], EndpointAuthSep)
	if len(strs) != 2 {
		return nil, fmt.Errorf("%v,%v", EndpointInvalid, endpoint)
	}
	user := strs[0]
	pass := strs[1]

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	/*
		Host:         "localhost:8332",
		User:         "yourrpcuser",
		Pass:         "yourrpcpass",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	*/
	connCfg := &rpcclient.ConnConfig{
		Host: host,
		User: user,
		Pass: pass,
		// default mainnet
		Params:       configParam[v],
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	cli, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	ok, err = WalletIsSync(cli)
	if !ok || err != nil {
		cli.Shutdown()
		return nil, fmt.Errorf("endpoint not sync %v", err)
	}

	return cli, nil
}

func WalletIsSync(cli *rpcclient.Client) (bool, error) {
	rets, err := cli.GetBlockChainInfo()
	if err != nil {
		return false, err
	}

	if rets.Headers < rets.Blocks {
		return false, fmt.Errorf(
			"wallet is not completed synchronization, current height %v, heightest height %v",
			rets.Headers, rets.Blocks)
	}

	return true, nil
}

func (bClients *BClients) WithClient(ctx context.Context, fn func(c *rpcclient.Client) (bool, error)) error {
	var (
		apiErr, err error
		retry       bool
		client      *rpcclient.Client
	)
	endpointmgr, err := endpoints.NewManager()
	if err != nil {
		return err
	}

	for i := 0; i < MaxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		client, err = bClients.GetNode(ctx, endpointmgr)

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

		client.Shutdown()
		if !retry {
			return apiErr
		}
	}
	return err
}

func Client() BClientI {
	return &BClients{}
}
