package api

import (
	"context"
	"net/http"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	pconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/fil"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) MpoolPush(ctx context.Context, in *sphinxplugin.MpoolPushRequest) (*sphinxplugin.MpoolPushResponse, error) {
	inMsg := in.GetMessage()
	inSign := in.GetSignature()

	to, err := address.NewFromString(inMsg.GetTo())
	if err != nil {
		logger.Sugar().Errorf("[%s] call NewFromString Addr: %s error: %v",
			pconst.FormatServiceName(),
			inMsg.GetTo(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			inMsg.GetTo(),
			err,
		)
	}

	from, err := address.NewFromString(inMsg.GetFrom())
	if err != nil {
		logger.Sugar().Errorf("[%s] call NewFromString Addr: %s error: %v",
			pconst.FormatServiceName(),
			inMsg.GetFrom(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			inMsg.GetFrom(),
			err,
		)
	}

	signType, err := fil.SignType(inSign.GetSignType())
	if err != nil {
		logger.Sugar().Errorf("[%s] call SignType SignType: %s error: %v",
			pconst.FormatServiceName(),
			inSign.GetSignType(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"Addr: %s error: %v",
			inSign.GetSignType(),
			err,
		)
	}
	signMsg := &types.SignedMessage{
		Message: types.Message{
			To:         to,
			From:       from,
			Method:     abi.MethodNum(inMsg.GetMethod()),
			Nonce:      inMsg.GetNonce(),
			Value:      abi.NewTokenAmount(1231243221000010),
			GasLimit:   655063,
			GasFeeCap:  abi.NewTokenAmount(2300),
			GasPremium: abi.NewTokenAmount(2250),
		},
		Signature: crypto.Signature{
			Type: signType,
			Data: inSign.GetData(),
		},
	}

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

	cid, err := api.MpoolPush(context.Background(), signMsg)
	if err != nil {
		logger.Sugar().Errorf("[%s] call MpoolPush From: %s To: %s Value: %v error: %v",
			pconst.FormatServiceName(),
			inMsg.GetFrom(),
			inMsg.GetTo(),
			inMsg.GetValue(),
			err,
		)
		return nil, status.Errorf(
			codes.Internal,
			"From: %s To: %s Value: %v error: %v",
			inMsg.GetFrom(),
			inMsg.GetTo(),
			inMsg.GetValue(),
			err,
		)
	}

	return &sphinxplugin.MpoolPushResponse{
		Info: &sphinxplugin.MpoolPushInfo{
			CID: cid.String(),
		},
	}, nil
}
