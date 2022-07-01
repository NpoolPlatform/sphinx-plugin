package plugin

import (
	"context"
	"net/http"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
)

var closer jsonrpc.ClientCloser

func Close() {
	if closer != nil {
		closer()
	}
	// api.Shutdown(context.Background())
}

func client() (v0api.FullNode, error) {
	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		return nil, env.ErrENVCoinTokenNotFound
	}

	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINLOCALAPI)
	if !ok {
		return nil, env.ErrENVCoinLocalAPINotFound
	}

	// if api != nil {
	// 	return api, nil
	// }

	var err error
	var api v0api.FullNodeStruct
	// internal has conn pool
	closer, err = jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", lotusapi.GetInternalStructs(&api), headers)
	if err != nil {
		return nil, err
	}

	return &api, nil
}
