package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	tronclient "github.com/Geapefurit/gotron-sdk/pkg/client"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"

	"github.com/Geapefurit/gotron-sdk/pkg/common"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/core"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

// here register plugin func
func init() {
	// // main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetron,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetron,
		sphinxproxy.TransactionState_TransactionStateWait,
		BuildTransaciton,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetron,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		BroadcastTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetron,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	// // test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypettron,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypettron,
		sphinxproxy.TransactionState_TransactionStateWait,
		BuildTransaciton,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypettron,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		BroadcastTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypettron,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetron, tron.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypettron, tron.TxFailErr)
	if err != nil {
		panic(err)
	}
}

// redefine Code ,because github.com/Geapefurit/gotron-sdk/pkg/proto/core/Tron.pb.go line 564 spelling err
const (
	TransactionInfoSUCCESS = 0
	TransactionInfoFAILED  = 1
)

func WalletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return in, err
	}

	if err := tron.ValidAddress(wbReq.Address); err != nil {
		return in, err
	}

	client := tron.Client()

	var bl int64
	err = client.WithClient(func(cli *tronclient.GrpcClient) (bool, error) {
		acc, err := cli.GetAccount(wbReq.Address)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			bl = tron.EmptyTRX
			return false, nil
		}
		if err != nil || acc == nil {
			return true, err
		}
		bl = acc.GetBalance()
		return false, nil
	})

	if err != nil {
		return in, err
	}

	wbResp := &ct.WalletBalanceResponse{}
	f := tron.TRXToBigFloat(bl)

	wbResp.Balance, _ = f.Float64()
	wbResp.BalanceStr = f.Text('f', tron.TRXACCURACY)

	return json.Marshal(wbResp)
}

func BuildTransaciton(ctx context.Context, in []byte) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return in, err
	}

	from := baseInfo.From
	to := baseInfo.To
	amount := tron.TRXToInt(baseInfo.Value)

	err = tron.ValidAddress(from)
	if err != nil {
		return in, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}
	err = tron.ValidAddress(to)
	if err != nil {
		return in, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	client := tron.Client()

	var txExtension *api.TransactionExtention
	err = client.WithClient(func(cli *tronclient.GrpcClient) (bool, error) {
		_, err := cli.GetAccount(from)
		if err != nil {
			return true, err
		}
		if tron.TxFailErr(err) {
			return false, err
		}
		txExtension, err = cli.Transfer(from, to, amount)
		if err != nil || txExtension == nil {
			return true, err
		}
		return false, err
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

func BroadcastTransaction(ctx context.Context, in []byte) (out []byte, err error) {
	bReq := &tron.BroadcastRequest{}
	err = json.Unmarshal(in, bReq)
	if err != nil {
		return in, err
	}

	client := tron.Client()
	transaction := bReq.TxExtension.Transaction
	bReq.TxExtension.GetTxid()
	var result *api.Return
	err = client.WithClient(func(cli *tronclient.GrpcClient) (bool, error) {
		result, err = cli.Broadcast(transaction)
		fmt.Println(result, err)
		if err != nil && result != nil && result.GetCode() == api.Return_TRANSACTION_EXPIRATION_ERROR {
			return false, err
		}
		if err != nil || result == nil {
			return true, err
		}
		return false, err
	})

	if err != nil {
		return in, err
	}
	if result == nil {
		return in, fmt.Errorf("get result faild")
	}

	if api.Return_SUCCESS == result.Code {
		bResp := &ct.BroadcastInfo{TxID: common.BytesToHexString(bReq.TxExtension.GetTxid())}
		if result.Result {
			return json.Marshal(bResp)
		}
	}

	failCodes := []api.ReturnResponseCode{
		// api.Return_SUCCESS,
		api.Return_SIGERROR,
		api.Return_CONTRACT_VALIDATE_ERROR,
		api.Return_CONTRACT_EXE_ERROR,
		// api.Return_BANDWIDTH_ERROR=4,
		4,
		api.Return_DUP_TRANSACTION_ERROR,
		api.Return_TAPOS_ERROR,
		api.Return_TOO_BIG_TRANSACTION_ERROR,
		api.Return_TRANSACTION_EXPIRATION_ERROR,
		// api.Return_SERVER_BUSY,
		// api.Return_NO_CONNECTION,
		// api.Return_NOT_ENOUGH_EFFECTIVE_CONNECTION,
		api.Return_OTHER_ERROR,
	}
	for _, v := range failCodes {
		if v == result.Code {
			return in, env.ErrTransactionFail
		}
	}

	return in, errors.New(string(result.GetMessage()))
}

// done(on chain) => true
func SyncTxState(ctx context.Context, in []byte) (out []byte, err error) {
	syncReq := &ct.SyncRequest{}
	err = json.Unmarshal(in, syncReq)
	if err != nil {
		return in, err
	}
	client := tron.Client()

	var txInfo *core.TransactionInfo
	err = client.WithClient(func(cli *tronclient.GrpcClient) (bool, error) {
		txInfo, err = cli.GetTransactionInfoByID(syncReq.TxID)
		if err != nil {
			return true, err
		}
		return false, err
	})

	if txInfo == nil || err != nil {
		return in, env.ErrWaitMessageOnChain
	}

	if txInfo.GetResult() != TransactionInfoSUCCESS {
		return in, env.ErrTransactionFail
	}

	if txInfo.Receipt.GetResult() != core.Transaction_Result_SUCCESS && txInfo.Receipt.GetResult() != core.Transaction_Result_DEFAULT {
		return in, env.ErrTransactionFail
	}

	syncResp := &ct.SyncResponse{ExitCode: 0}
	return json.Marshal(syncResp)
}
