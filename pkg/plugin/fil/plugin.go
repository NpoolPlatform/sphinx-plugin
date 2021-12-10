package fil

import (
	"context"
	"errors"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
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
	val, err := types.ParseFIL(decimal.NewFromFloat(inMsg.GetValue()).String())
	if err != nil {
		return "", err
	}
	signMsg := &types.SignedMessage{
		Message: types.Message{
			To:         to,
			From:       from,
			Method:     abi.MethodNum(inMsg.GetMethod()),
			Nonce:      inMsg.GetNonce(),
			Value:      abi.TokenAmount(val),
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

func StateSearchMsg(_ctx context.Context, in *sphinxproxy.ProxyPluginRequest) (*lotusapi.MsgLookup, error) {
	api, closer, err := client()
	if err != nil {
		return nil, err
	}
	defer closer()

	_cid, err := cid.Decode(in.GetCID())
	if err != nil {
		return nil, ErrCIDInvalid
	}

	if err := waitMessageOut(api, _cid); err != nil {
		return nil, err
	}
	return waitMessageOnChain(api, _cid)
}

// wait message on out
func waitMessageOut(api v0api.FullNode, _cid cid.Cid) error {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.WaitMsgOutTimeout)
	defer cancel()
	for {
		select {
		// 40 seconds timeout possible gas too low
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			mp, err := api.MpoolPending(ctx, types.EmptyTSK)
			if err != nil {
				return err
			}
			if !includeCID(_cid, mp) {
				return nil
			}
		}
	}
}

// wait message on chain
func waitMessageOnChain(api v0api.FullNode, _cid cid.Cid) (*lotusapi.MsgLookup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sconst.WaitMsgOutTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			// TODO double-spend
			chainMsg, err := api.StateSearchMsg(ctx, _cid)
			if err != nil {
				return chainMsg, err
			}
			if chainMsg != nil {
				return chainMsg, nil
			}
		}
	}
}

func includeCID(_cid cid.Cid, sms []*types.SignedMessage) bool {
	for _, mCid := range sms {
		if mCid.Cid() == _cid {
			return true
		}
	}
	return false
}
