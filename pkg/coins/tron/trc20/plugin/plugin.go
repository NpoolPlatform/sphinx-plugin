package trc20

import (
	"context"
	"encoding/json"
	"math/big"
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
	bl, err := BalanceS(wbReq.Address, contract)
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

	txExtension, err := BuildTransacitonS(
		baseInfo.From,
		baseInfo.To,
		contract,
		tron.TRC20ToBigInt(baseInfo.Value),
		tron.TRC20FeeLimit,
	)
	if err != nil {
		return in, err
	}

	signTx := &tron.SignMsgTx{
		Base:        *baseInfo,
		TxExtension: txExtension,
	}

	return json.Marshal(signTx)
}

func BalanceS(addr, contractAddress string) (*big.Int, error) {
	var err error
	ret := tron.EmptyTRC20
	if err := tron.ValidAddress(addr); err != nil {
		return ret, err
	}

	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		ret, err = c.TRC20ContractBalance(addr, contractAddress)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			ret = tron.EmptyTRC20
			return false, nil
		}
		return true, err
	})

	return ret, err
}

func BuildTransacitonS(from, to, contract string, amount *big.Int, feeLimit int64) (*api.TransactionExtention, error) {
	var ret *api.TransactionExtention
	var err error
	client := tron.Client()
	err = client.WithClient(func(c *tronclient.GrpcClient) (bool, error) {
		ret, err = c.TRC20Send(from, to, contract, amount, feeLimit)
		return true, err
	})
	return ret, err
}
