package trc20

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/plugin/tron"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

var (
	// ErrWaitMessageOnChain ..
	ErrWaitMessageOnChain = errors.New("wait message on chain")
	// ErrAddrNotValid ..
	ErrAddrNotValid = errors.New("invalid address")
	// ErrTransactionFail ..
	ErrTransactionFail = errors.New("transaction fail")
)

func WalletBalance(ctx context.Context, wallet string) (balance *big.Int, err error) {
	contractID, ok := env.LookupEnv(env.ENVCONTRACTID)
	if !ok {
		return EmptyInt, env.ErrENVContractIDNotFound
	}

	client, err := tron.Client()
	if err != nil {
		return EmptyInt, err
	}
	return client.TRC20ContractBalance(wallet, contractID)
}

func TransactionSend(ctx context.Context, req *sphinxproxy.ProxyPluginRequest) (*api.TransactionExtention, error) {
	contractID, ok := env.LookupEnv(env.ENVCONTRACTID)
	if !ok {
		return nil, env.ErrENVContractIDNotFound
	}

	from := req.GetMessage().GetFrom()
	to := req.GetMessage().GetTo()
	amount := req.GetMessage().GetValue()
	fee := feeLimit

	client, err := tron.Client()
	if err != nil {
		return nil, err
	}
	return client.TRC20Send(from, to, contractID, ToInt(amount), fee)
}

func BroadcastTransaction(ctx context.Context, transaction *core.Transaction) (err error) {
	client, err := tron.Client()
	if err != nil {
		return err
	}

	client.SetTimeout(10 * time.Second)
	result, err := client.Broadcast(transaction)
	if err != nil {
		return err
	}
	if result.Code != 0 {
		return errors.New(string(result.GetMessage()))
	}
	if result.Result {
		return nil
	}
	return errors.New(string(result.GetMessage()))
}

// done(on chain) => true
func SyncTxState(ctx context.Context, cid string) (bool, error) {
	client, err := tron.Client()
	if err != nil {
		return false, err
	}

	txInfo, err := client.GetTransactionInfoByID(cid)
	if err != nil {
		return false, err
	}
	if txInfo == nil {
		return false, ErrWaitMessageOnChain
	}

	if txInfo.GetResult() != 0 {
		return false, ErrTransactionFail
	}

	if txInfo.Receipt.GetResult() == 1 {
		return true, nil
	}
	return false, ErrTransactionFail
}
