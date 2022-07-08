package plugin

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/tron"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
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

	// err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetron, bsc.TxFailErr)
	// if err != nil {
	// 	panic(err)
	// }

	// err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypettron, bsc.TxFailErr)
	// if err != nil {
	// 	panic(err)
	// }

	// coins.RegisterAbortErr(
	// 	bsc.ErrTransactionFail,
	// 	bsc.ErrAddrNotValid,
	// )
}

// redefine Code ,because github.com/fbsobreira/gotron-sdk/pkg/proto/core/Tron.pb.go line 564 spelling err
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
	client := tron.Client()

	bl, err := client.TRXBalanceS(wbReq.Address)
	if err != nil {
		return in, err
	}

	wbResp := &ct.WalletBalanceResponse{}
	f := tron.TRXToBigFloat(bl)

	wbResp.Balance, _ = f.Float64()
	wbResp.BalanceStr = f.Text('f', tron.TRXACCURACY)

	out, err = json.Marshal(wbResp)
	return out, err
}

func BuildTransaciton(ctx context.Context, in []byte) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return in, err
	}
	from := baseInfo.From
	to := baseInfo.To
	amount := baseInfo.Value

	client := tron.Client()

	txExtension, err := client.TRXTransferS(from, to, tron.TRXToInt(amount))
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
	result, err := client.BroadcastS(transaction)
	if err != nil {
		return in, err
	}
	if result.Code != 0 {
		return in, errors.New(string(result.GetMessage()))
	}

	bResp := &ct.BroadcastInfo{TxID: string(bReq.TxExtension.GetTxid())}

	if result.Result {
		return json.Marshal(bResp)
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

	txInfo, err := client.GetTransactionInfoByIDS(syncReq.TxID)

	if txInfo == nil || err != nil {
		return in, tron.ErrWaitMessageOnChain
	}

	logger.Sugar().Infof("transaction info {CID: %v ,ChainResult: %v, TxResult: %v, Fee: %v }", syncReq.TxID, txInfo.GetResult(), txInfo.GetReceipt().GetResult(), txInfo.GetFee())

	if txInfo.GetResult() != TransactionInfoSUCCESS {
		return in, tron.ErrTransactionFail
	}

	if txInfo.Receipt.GetResult() != core.Transaction_Result_SUCCESS && txInfo.Receipt.GetResult() != core.Transaction_Result_DEFAULT {
		return in, tron.ErrTransactionFail
	}

	syncResp := &ct.SyncResponse{ExitCode: 0}
	return json.Marshal(syncResp)
}
