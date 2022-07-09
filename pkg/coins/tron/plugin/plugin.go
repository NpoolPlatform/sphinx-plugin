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
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
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

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetron, tron.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypettron, tron.TxFailErr)
	if err != nil {
		panic(err)
	}

	coins.RegisterAbortErr(
		tron.ErrTransactionFail,
		tron.ErrInvalidAddr,
		tron.ErrAddressEmpty,
	)
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
	err = tron.ValidAddress(from)
	if err != nil {
		return in, err
	}
	err = tron.ValidAddress(to)
	if err != nil {
		return in, err
	}
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
	bReq.TxExtension.GetTxid()
	result, err := client.BroadcastS(transaction)
	if err != nil {
		return in, err
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
			return in, tron.ErrTransactionFail
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
