package api

import (
	"context"
	"net/http"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) MpoolGetNonce(ctx context.Context, in *sphinxplugin.MpoolGetNonceRequest) (*sphinxplugin.MpoolGetNonceResponse, error) {
	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		logger.Sugar().Errorf("[%s] call LookupEnv ENV: %s not found",
			pconst.FormatServiceName(),
			env.ENVCOINTOKEN,
		)
		return nil, status.Errorf(codes.Internal, "ENV: %s not found", env.ENVCOINTOKEN)
	}
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		logger.Sugar().Errorf("[%s] call LookupEnv ENV: %s not found",
			pconst.FormatServiceName(),
			env.ENVCOINAPI,
		)
		return nil, status.Errorf(codes.Internal, "ENV: %s not found", env.ENVCOINAPI)
	}

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		logger.Sugar().Errorf("[%s] call NewMergeClient Addr: %s error: %v",
			pconst.FormatServiceName(),
			addr,
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			addr,
			err,
		)
	}
	defer closer()

	from, err := address.NewFromString(in.GetAddress())
	if err != nil {
		logger.Sugar().Errorf("[%s] call NewFromString Addr: %s error: %v",
			pconst.FormatServiceName(),
			in.GetAddress(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			in.GetAddress(),
			err,
		)
	}

	nonce, err := api.MpoolGetNonce(context.Background(), from)
	if err != nil {
		logger.Sugar().Errorf("[%s] call NewFromString Addr: %s error: %v",
			pconst.FormatServiceName(),
			in.GetAddress(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			in.GetAddress(),
			err,
		)
	}
	return &sphinxplugin.MpoolGetNonceResponse{
		Info: &sphinxplugin.MpoolGetNonceInfo{
			Nonce: nonce,
		},
	}, nil
}
