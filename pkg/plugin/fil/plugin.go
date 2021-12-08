package fil

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

var (
	ErrAddressInvalid       = errors.New("address invalid")
	ErrENVCoinTokenNotFound = errors.New("env ENV_COIN_TOKEN not found")
	ErrENVCoinAPINotFound   = errors.New("env ENV_COIN_API not found")
	ErrSignTypeInvalid      = errors.New("sign type invalid")
	ErrFindMsgNotFound      = errors.New("failed to find message")
	ErrCIDInvalid           = errors.New("cid invalid")
)

func WalletBalance(ctx context.Context, wallet string) (balance types.BigInt, err error) {
	if wallet == "" {
		return types.EmptyInt, ErrAddressInvalid
	}

	from, err := address.NewFromString(wallet)
	if err != nil {
		return types.EmptyInt, err
	}

	api, closer, err := client()
	if err != nil {
		return types.EmptyInt, err
	}
	defer closer()

	return api.WalletBalance(ctx, from)
}

func MpoolGetNonce(ctx context.Context, wallet string) (nonce uint64, err error) {
	if wallet == "" {
		return 0, ErrAddressInvalid
	}

	from, err := address.NewFromString(wallet)
	if err != nil {
		return 0, err
	}

	api, closer, err := client()
	if err != nil {
		return 0, err
	}
	defer closer()

	_nonce, err := api.MpoolGetNonce(ctx, from)
	if err != nil {
		return 0, err
	}

	return _nonce, nil
}

func MpoolPush(ctx context.Context, inMsg *sphinxplugin.UnsignedMessage, inSign *sphinxplugin.Signature) (chainID string, err error) {
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

	api, closer, err := client()
	if err != nil {
		return "", err
	}
	defer closer()

	_cid, err := api.MpoolPush(ctx, signMsg)
	if err != nil {
		return "", err
	}

	return _cid.String(), nil
}

func StateSearchMsg(ctx context.Context, in *sphinxproxy.ProxyPluginRequest) (*lotusapi.MsgLookup, error) {
	api, closer, err := client()
	if err != nil {
		return nil, err
	}
	defer closer()

	_cid, err := cid.Decode(in.GetCID())
	if err != nil {
		return nil, ErrCIDInvalid
	}
	chainMsg, err := api.StateSearchMsg(ctx, _cid)
	if err != nil {
		return nil, err
	}
	if chainMsg == nil {
		return nil, ErrFindMsgNotFound
	}

	return chainMsg, nil
}
