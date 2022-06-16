package tron

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
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

// redefine Code ,because github.com/fbsobreira/gotron-sdk/pkg/proto/core/Tron.pb.go line 564 spelling err
const (
	TransactionInfoSUCCESS = 0
	TransactionInfoFAILED  = 1
)

func WalletBalance(ctx context.Context, wallet string) (balance int64, err error) {
	client := Client()
	return client.TRXBalanceS(wallet)
}

func Transfer(ctx context.Context, req *sphinxproxy.ProxyPluginRequest) (*api.TransactionExtention, error) {
	from := req.GetMessage().GetFrom()
	to := req.GetMessage().GetTo()
	amount := req.GetMessage().GetValue()

	client := Client()

	return client.TRXTransferS(from, to, TRXToInt(amount))
}

func BroadcastTransaction(ctx context.Context, transaction *core.Transaction) (err error) {
	client := Client()

	result, err := client.BroadcastS(transaction)
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
func SyncTxState(ctx context.Context, cid string) (pending bool, exitcode int64, err error) {
	client := Client()

	txInfo, err := client.GetTransactionInfoByIDS(cid)

	if txInfo == nil || err != nil {
		return false, 0, ErrWaitMessageOnChain
	}

	logger.Sugar().Infof("transaction info {CID: %v ,ChainResult: %v, TxResult: %v, Fee: %v }", cid, txInfo.GetResult(), txInfo.GetReceipt().GetResult(), txInfo.GetFee())

	if txInfo.GetResult() != TransactionInfoSUCCESS {
		return true, TransactionInfoFAILED, fmt.Errorf("trasction fail, %v, %v", txInfo.GetResult(), txInfo.GetReceipt().GetResult())
	}

	if txInfo.Receipt.GetResult() != core.Transaction_Result_SUCCESS && txInfo.Receipt.GetResult() != core.Transaction_Result_DEFAULT {
		return true, TransactionInfoFAILED, fmt.Errorf("trasction fail, %v", txInfo.GetReceipt().GetResult())
	}

	return true, 0, nil
}
