package trc20

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tronclient "github.com/Geapefurit/gotron-sdk/pkg/client"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	tron_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

// here register plugin func
func init() {
	// // main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		BuildTransaciton,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		tron_plugin.BroadcastTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeusdttrc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		tron_plugin.SyncTxState,
	)

	// // test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionState_TransactionStateWait,
		BuildTransaciton,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		tron_plugin.BroadcastTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetusdttrc20,
		sphinxproxy.TransactionState_TransactionStateSync,
		tron_plugin.SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeusdttrc20, tron.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetusdttrc20, tron.TxFailErr)
	if err != nil {
		panic(err)
	}
}

func WalletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}

	v, ok := env.LookupEnv(env.ENVCOINNET)
	if !ok {
		return nil, env.ErrEVNCoinNet
	}
	if !coins.CheckSupportNet(v) {
		return nil, env.ErrEVNCoinNetValue
	}

	contract := tron.USDTContract(v)
	err = tron.ValidAddress(contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", contract, tron.AddressInvalid, err)
	}

	bl := tron.EmptyTRC20
	if err := tron.ValidAddress(wbReq.Address); err != nil {
		return nil, err
	}

	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		bl, err = c.TRC20ContractBalance(wbReq.Address, contract)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			bl = tron.EmptyTRC20
			return false, nil
		}
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	f := tron.TRC20ToBigFloat(bl)
	wbResp := &ct.WalletBalanceResponse{}

	wbResp.Balance, _ = f.Float64()
	wbResp.BalanceStr = f.Text('f', tron.TRC20ACCURACY)

	out, err = json.Marshal(wbResp)

	return out, err
}

func BuildTransaciton(ctx context.Context, in []byte) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return nil, err
	}

	if !coins.CheckSupportNet(baseInfo.ENV) {
		return nil, env.ErrEVNCoinNetValue
	}

	err = tron.ValidAddress(baseInfo.From)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	err = tron.ValidAddress(baseInfo.To)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	contract := tron.USDTContract(baseInfo.ENV)
	err = tron.ValidAddress(contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", contract, tron.AddressInvalid, err)
	}

	var txExtension *api.TransactionExtention
	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		txExtension, err = c.TRC20Send(
			baseInfo.From,
			baseInfo.To,
			contract,
			tron.TRC20ToBigInt(baseInfo.Value),
			tron.TRC20FeeLimit,
		)
		return false, err
	})
	if err != nil {
		return nil, err
	}
	signTx := &tron.SignMsgTx{
		Base:        *baseInfo,
		TxExtension: txExtension,
	}

	return json.Marshal(signTx)
}
