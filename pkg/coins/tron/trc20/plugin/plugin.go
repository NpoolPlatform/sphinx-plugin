package trc20

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	tron_plugin "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron/plugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
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
		return in, err
	}
	contract := config.GetENV().Contract

	bl := tron.EmptyTRC20
	if err := tron.ValidAddress(wbReq.Address); err != nil {
		return in, err
	}

	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		bl, err = c.TRC20ContractBalance(wbReq.Address, contract)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			bl = tron.EmptyTRC20
			return false, nil
		}
		return true, err
	})
	if err != nil {
		return in, err
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
		return in, err
	}

	contract := config.GetENV().Contract

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
		return true, err
	})
	if err != nil {
		return in, err
	}
	signTx := &tron.SignMsgTx{
		Base:        *baseInfo,
		TxExtension: txExtension,
	}

	return json.Marshal(signTx)
}
