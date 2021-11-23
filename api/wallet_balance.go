package api

// import (
// 	"context"
// 	"net/http"

// 	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
// 	"github.com/NpoolPlatform/message/npool/sphinxplugin"
// 	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
// 	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
// 	"github.com/filecoin-project/go-address"
// 	"github.com/filecoin-project/go-jsonrpc"
// 	lotusapi "github.com/filecoin-project/lotus/api"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func (s *Server) WalletBalance(ctx context.Context, in *sphinxplugin.WalletBalanceRequest) (*sphinxplugin.WalletBalanceResponse, error) {
// 	if in.GetAddress() == "" {
// 		logger.Sugar().Errorf("[%s] check Address:%v empty",
// 			pconst.FormatServiceName(),
// 			in.GetAddress(),
// 		)
// 		return nil, status.Errorf(
// 			codes.InvalidArgument,
// 			"Address:%v empty",
// 			in.GetAddress(),
// 		)
// 	}

// 	from, err := address.NewFromString(in.GetAddress())
// 	if err != nil {
// 		logger.Sugar().Errorf("[%s] call NewFromString Address: %v error: %v",
// 			pconst.FormatServiceName(),
// 			in.GetAddress(),
// 			err,
// 		)
// 		return nil, status.Errorf(
// 			codes.InvalidArgument,
// 			"Address:%v not valid",
// 			in.GetAddress(),
// 		)
// 	}

// 	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
// 	if !ok {
// 		logger.Sugar().Errorf("[%s] get lotus api env: %v not found",
// 			pconst.FormatServiceName(),
// 			env.ENVCOINTOKEN,
// 		)
// 		return nil, status.Error(
// 			codes.Internal,
// 			"internal server error",
// 		)
// 	}
// 	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

// 	addr, ok := env.LookupEnv(env.ENVCOINAPI)
// 	if !ok {
// 		logger.Sugar().Errorf("[%s] check CoinType:%v error: %v",
// 			pconst.FormatServiceName(),
// 			in.GetCoinType(),
// 			err,
// 		)
// 		return nil, status.Error(
// 			codes.Internal,
// 			"internal server error",
// 		)
// 	}

// 	var api lotusapi.FullNodeStruct
// 	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
// 	if err != nil {
// 		logger.Sugar().Errorf("[%s] call NewMergeClient Addr: %s error: %v",
// 			pconst.FormatServiceName(),
// 			addr,
// 			err,
// 		)
// 		return nil, status.Errorf(
// 			codes.Internal,
// 			"Addr: %s error: %v",
// 			addr,
// 			err,
// 		)
// 	}
// 	defer closer()

// 	balance, err := api.WalletBalance(context.Background(), from)
// 	if err != nil {
// 		logger.Sugar().Errorf("[%s] call WalletBalance Addr: %v error: %v",
// 			pconst.FormatServiceName(),
// 			in.GetAddress(),
// 			err,
// 		)
// 		return nil, status.Errorf(
// 			codes.Internal,
// 			"Addr: %v error: %v",
// 			in.GetAddress(),
// 			err,
// 		)
// 	}
// 	return &sphinxplugin.WalletBalanceResponse{
// 		Info: &sphinxplugin.WalletBalanceInfo{
// 			Balance: balance.String(),
// 		},
// 	}, nil
// }
