package fil

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/ipfs/go-cid"

	"github.com/shopspring/decimal"
)

// here register plugin func
func init() {
	// main
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxplugin.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxplugin.TransactionType_PreSign,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxplugin.TransactionType_Broadcast,
		Broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxplugin.TransactionType_SyncMsgState,
		SyncTx,
	)

	// test
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxplugin.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxplugin.TransactionType_PreSign,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxplugin.TransactionType_Broadcast,
		Broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxplugin.TransactionType_SyncMsgState,
		SyncTx,
	)
}

func WalletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}
	// TODO in main init
	address.CurrentNetwork = coins.FILNetMap[coins.CoinNet]

	if info.Address == "" {
		return nil, env.ErrAddressInvalid
	}

	from, err := address.NewFromString(info.Address)
	if err != nil {
		return nil, err
	}

	api, err := client()
	if err != nil {
		return nil, err
	}

	chainBalance, err := api.WalletBalance(ctx, from)
	if err != nil {
		return nil, err
	}

	balance, ok := big.NewFloat(0).SetString(chainBalance.String())
	if !ok {
		return nil, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(float64((build.FilecoinPrecision))))
	f, exact := balance.Float64()
	if exact != big.Exact {
		logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	_out := ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}

	return json.Marshal(_out)
}

func PreSign(ctx context.Context, in []byte) (out []byte, err error) {
	info := PreSignRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	// TODO in main init
	address.CurrentNetwork = coins.FILNetMap[coins.CoinNet]

	if info.Address == "" {
		return nil, env.ErrAddressInvalid
	}

	from, err := address.NewFromString(info.Address)
	if err != nil {
		return nil, err
	}

	api, err := client()
	if err != nil {
		return nil, err
	}

	_nonce, err := api.MpoolGetNonce(ctx, from)
	if err != nil {
		return nil, err
	}

	_out := PreSignReponse{
		Info: RawTx{
			GasLimit:   200000000,
			GasFeeCap:  10000000,
			GasPremium: 1000000,
			Method:     uint64(builtin.MethodSend),
			Nonce:      _nonce,
		},
	}

	return json.Marshal(_out)
}

func Broadcast(ctx context.Context, in []byte) (out []byte, err error) {
	info := BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	raw := info.Raw
	signed := info.Signature

	// TODO in main init
	address.CurrentNetwork = coins.FILNetMap[coins.CoinNet]

	to, err := address.NewFromString(raw.To)
	if err != nil {
		return nil, env.ErrAddressInvalid
	}

	from, err := address.NewFromString(raw.From)
	if err != nil {
		return nil, env.ErrAddressInvalid
	}

	signType, err := SignType(signed.SignType)
	if err != nil {
		return nil, env.ErrSignTypeInvalid
	}
	val, err := types.ParseFIL(decimal.NewFromFloat(raw.Value).String())
	if err != nil {
		return nil, err
	}
	signMsg := &types.SignedMessage{
		Message: types.Message{
			To:         to,
			From:       from,
			Method:     abi.MethodNum(raw.Method),
			Nonce:      raw.Nonce,
			Value:      abi.TokenAmount(val),
			GasLimit:   raw.GasLimit,
			GasFeeCap:  abi.NewTokenAmount(raw.GasFeeCap),
			GasPremium: abi.NewTokenAmount(raw.GasPremium),
		},
		Signature: crypto.Signature{
			Type: signType,
			Data: signed.Data,
		},
	}

	api, err := client()
	if err != nil {
		return nil, err
	}

	_cid, err := api.MpoolPush(ctx, signMsg)
	if err != nil {
		return nil, err
	}

	_out := BroadcastResponse{
		TxID: _cid.String(),
	}

	return json.Marshal(_out)
}

func SyncTx(_ctx context.Context, in []byte) (out []byte, err error) {
	info := SyncTxRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	// TODO in main init
	address.CurrentNetwork = coins.FILNetMap[coins.CoinNet]

	api, err := client()
	if err != nil {
		return nil, err
	}

	_cid, err := cid.Decode(info.TxID)
	if err != nil {
		return nil, env.ErrCIDInvalid
	}

	if err := waitMessageOut(api, _cid); err != nil {
		return nil, err
	}

	msgLookUP, err := waitMessageOnChain(api, _cid)
	if err != nil {
		return nil, err
	}

	_out := SyncTxResponse{
		ExitCode: int64(msgLookUP.Receipt.ExitCode),
	}

	return json.Marshal(_out)
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