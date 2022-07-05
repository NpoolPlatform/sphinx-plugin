package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/fil"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	sconst "github.com/NpoolPlatform/sphinx-plugin/pkg/message/const"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/ipfs/go-cid"

	"github.com/shopspring/decimal"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypefilecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		preSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		broadcast,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetfilecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		syncTx,
	)

	// register err fsm
	coins.RegisterAbortErr(
		env.ErrEVNCoinNet,
		env.ErrEVNCoinNetValue,
		env.ErrAddressInvalid,
		env.ErrSignTypeInvalid,
		env.ErrCIDInvalid,
	)

	// register err not a value handle
	// coins.RegisterAbortFuncErr(
	// 	sphinxplugin.CoinType_CoinTypefilecoin,
	// 	func(err error) bool {
	// 		return true
	// 	})
	// coins.RegisterAbortFuncErr(
	// 	sphinxplugin.CoinType_CoinTypetfilecoin,
	// 	func(err error) bool {
	// 		return false
	// 	})
}

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.WalletBalanceRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	// TODO in main init
	address.CurrentNetwork = fil.FILNetMap[v]

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

	balance.Quo(balance, big.NewFloat(0).SetUint64(build.FilecoinPrecision))
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

func preSign(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.BaseInfo{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	// TODO in main init
	address.CurrentNetwork = fil.FILNetMap[info.ENV]

	if info.From == "" {
		return nil, env.ErrAddressInvalid
	}

	from, err := address.NewFromString(info.From)
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

	_out := fil.SignRequest{
		ENV: info.ENV,
		Info: fil.RawTx{
			To:         info.To,
			From:       info.From,
			Value:      info.Value,
			GasLimit:   200000000,
			GasFeeCap:  10000000,
			GasPremium: 1000000,
			Method:     uint64(builtin.MethodSend),
			Nonce:      _nonce,
		},
	}

	return json.Marshal(_out)
}

func broadcast(ctx context.Context, in []byte) (out []byte, err error) {
	info := fil.BroadcastRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	raw := info.Raw
	signed := info.Signature

	if !coins.CheckSupportNet(info.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	// TODO in main init
	address.CurrentNetwork = fil.FILNetMap[info.ENV]

	to, err := address.NewFromString(raw.To)
	if err != nil {
		return nil, env.ErrAddressInvalid
	}

	from, err := address.NewFromString(raw.From)
	if err != nil {
		return nil, env.ErrAddressInvalid
	}

	signType, err := fil.SignType(signed.SignType)
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

	_out := ct.SyncRequest{
		TxID: _cid.String(),
	}

	return json.Marshal(_out)
}

func syncTx(ctx context.Context, in []byte) (out []byte, err error) {
	info := ct.SyncRequest{}
	if err := json.Unmarshal(in, &info); err != nil {
		return nil, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	// TODO in main init
	address.CurrentNetwork = fil.FILNetMap[v]

	api, err := client()
	if err != nil {
		return nil, err
	}

	_cid, err := cid.Decode(info.TxID)
	if err != nil {
		return nil, env.ErrCIDInvalid
	}

	ctx, cancel := context.WithTimeout(ctx, sconst.WaitMsgOutTimeout)
	defer cancel()

	// 1. check message out
	mp, err := api.MpoolPending(ctx, types.EmptyTSK)
	if err != nil {
		return
	}
	if !includeCID(_cid, mp) {
		return
	}

	// 2. check message on chain
	chainMsg, err := api.StateSearchMsg(ctx, _cid)
	if err != nil {
		return nil, err
	}

	// if message not on chain chainMsg is nil, until wait message on chain
	if chainMsg == nil {
		return nil, env.ErrWaitMessageOnChain
	}

	// TODO: check message is replaced ?
	// chainMsg.Receipt.ExitCode != exitcode.Ok

	// check message on chain done
	_out := ct.SyncResponse{
		ExitCode: int64(chainMsg.Receipt.ExitCode),
	}

	return json.Marshal(_out)
}

/*
// waitMessageOut wait message on out
func waitMessageOut(ctx context.Context, api v0api.FullNode, _cid cid.Cid) error {
	var (
		errExit    = make(chan error)
		waitMsgOut = make(chan struct{})
	)

	// wait message out
	go func() {
		for {
			select {
			// 40 seconds timeout possible gas too low
			case <-ctx.Done():
				errExit <- ctx.Err()
				return
			case <-time.After(5 * time.Second):
				mp, err := api.MpoolPending(ctx, types.EmptyTSK)
				if err != nil {
					errExit <- ctx.Err()
					return
				}
				if !includeCID(_cid, mp) {
					waitMsgOut <- struct{}{}
					return
				}
			}
		}
	}()

	select {
	case _ = <-errExit:
	case <-waitMsgOut:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				// TODO double-spend
				chainMsg, err := api.StateSearchMsg(ctx, _cid)
				if err != nil {
					return err
				}
				if chainMsg != nil && chainMsg.Receipt.ExitCode != exitcode.Ok {}
			}
		}
	}
	return nil
}
*/

func includeCID(_cid cid.Cid, sms []*types.SignedMessage) bool {
	for _, mCid := range sms {
		if mCid.Cid() == _cid {
			return true
		}
	}
	return false
}
