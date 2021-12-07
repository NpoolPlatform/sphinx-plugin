package fil

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
)

var (
	ErrAddressInvalid       = errors.New("address invalid")
	ErrENVCoinTokenNotFound = errors.New("env ENV_COIN_TOKEN not found")
	ErrENVCoinAPINotFound   = errors.New("env ENV_COIN_API not found")
	ErrSignTypeInvalid      = errors.New("sign type invalid")
)

func WalletBalance(wallet string) (balance types.BigInt, err error) {
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

	return api.WalletBalance(context.Background(), from)
}

func MpoolGetNonce(wallet string) (nonce uint64, err error) {
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

	api, closer, err := client()
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
