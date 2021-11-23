package fil

import (
	"context"
	"errors"
	"net/http"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
)

var (
	ErrAddressInvalid       = errors.New("address invalid")
	ErrENVCoinTokenNotFound = errors.New("env ENV_COIN_TOKEN not found")
	ErrENVCoinAPINotFound   = errors.New("env ENV_COIN_API not found")
	ErrSignTypeInvalid      = errors.New("sign type invalid")
)

func WalletBalance(wallet string) (balance uint64, err error) {
	if wallet == "" {
		return 0, ErrAddressInvalid
	}

	from, err := address.NewFromString(wallet)
	if err != nil {
		return 0, err
	}

	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		return 0, ErrENVCoinTokenNotFound
	}
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return 0, ErrENVCoinAPINotFound
	}

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		return 0, err
	}
	defer closer()

	_balance, err := api.WalletBalance(context.Background(), from)
	if err != nil {
		return 0, err
	}

	return _balance.Uint64(), nil
}

func MpoolGetNonce(wallet string) (nonce uint64, err error) {
	if wallet == "" {
		return 0, ErrAddressInvalid
	}

	from, err := address.NewFromString(wallet)
	if err != nil {
		return 0, err
	}

	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		return 0, ErrENVCoinTokenNotFound
	}
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return 0, ErrENVCoinAPINotFound
	}

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		return 0, err
	}
	defer closer()

	_nonce, err := api.MpoolGetNonce(context.Background(), from)
	if err != nil {
		return 0, err
	}

	return _nonce, nil
}

func MpoolPush(inMsg *sphinxplugin.UnsignedMessage, inSign *sphinxplugin.Signature) (chainID string, err error) {
	to, err := address.NewFromString(inMsg.GetTo())
	if err != nil {
		return "", ErrAddressInvalid
	}

	from, err := address.NewFromString(inMsg.GetFrom())
	if err != nil {
		return "", ErrAddressInvalid
	}

	signType, err := SignType(inSign.GetSignType())
	if err != nil {
		return "", ErrSignTypeInvalid
	}
	signMsg := &types.SignedMessage{
		Message: types.Message{
			To:         to,
			From:       from,
			Method:     abi.MethodNum(inMsg.GetMethod()),
			Nonce:      inMsg.GetNonce(),
			Value:      abi.NewTokenAmount(int64(inMsg.GetValue())),
			GasLimit:   inMsg.GetGasLimit(),
			GasFeeCap:  abi.NewTokenAmount(int64(inMsg.GetGasFeeCap())),
			GasPremium: abi.NewTokenAmount(int64(inMsg.GetGasPremium())),
		},
		Signature: crypto.Signature{
			Type: signType,
			Data: inSign.GetData(),
		},
	}

	authToken, ok := env.LookupEnv(env.ENVCOINTOKEN)
	if !ok {
		return "", ErrENVCoinTokenNotFound
	}
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}

	addr, ok := env.LookupEnv(env.ENVCOINAPI)
	if !ok {
		return "", ErrENVCoinAPINotFound
	}

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		return "", err
	}
	defer closer()

	cid, err := api.MpoolPush(context.Background(), signMsg)
	if err != nil {
		return "", err
	}

	return cid.String(), nil
}
