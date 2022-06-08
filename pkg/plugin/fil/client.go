package fil

import (
	"context"
	"net/http"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
)

func client() (v0api.FullNode, jsonrpc.ClientCloser, error) {
	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		return nil, nil, env.ErrENVCoinTokenNotFound
	}
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINLOCALAPI)
	if !ok {
		return nil, nil, env.ErrENVCoinLocalAPINotFound
	}

	var api v0api.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", lotusapi.GetInternalStructs(&api), headers)
	if err != nil {
		return nil, nil, err
	}

	return &api, closer, nil
}
