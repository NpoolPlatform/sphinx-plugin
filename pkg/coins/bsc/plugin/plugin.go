package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	bsc "github.com/NpoolPlatform/sphinx-plugin/pkg/coins/bsc"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	plugin_types "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rlp"
)

// here register plugin func
func init() {
	// main
	// coins.RegisterBalance(
	// 	sphinxplugin.CoinType_CoinTypebinancecoin,
	// 	sphinxproxy.TransactionType_Balance,
	// 	WalletBalance,
	// )
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypebinancecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	// test
	// coins.RegisterBalance(
	// 	sphinxplugin.CoinType_CoinTypetbinancecoin,
	// 	sphinxproxy.TransactionType_Balance,
	// 	WalletBalance,
	// )
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetbinancecoin,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)
}

var (
	// ErrWaitMessageOnChain ..
	ErrWaitMessageOnChain = errors.New("wait message on chain")
	// ErrAddrNotValid ..
	ErrAddrNotValid = errors.New("invalid address")
	// ErrTransactionFail ..
	ErrTransactionFail = errors.New("transaction fail")
)

type PreSignInfo struct {
	ChainID    int64
	ContractID string
	Nonce      uint64
	GasPrice   int64
	GasLimit   int64
}

func WalletBalance(ctx context.Context, addr string) (*big.Int, error) {
	client := bsc.Client()

	if !common.IsHexAddress(addr) {
		return nil, ErrAddrNotValid
	}
	return client.BalanceAtS(ctx, common.HexToAddress(addr), nil)
}

func PreSign(ctx context.Context, in []byte) (out []byte, err error) {
	txRequest := &plugin_types.CreateTransactionRequest{}
	err = json.Unmarshal(in, txRequest)
	if err != nil {
		return nil, err
	}
	client := bsc.Client()

	if !common.IsHexAddress(txRequest.From) {
		return nil, ErrAddrNotValid
	}

	chainID, err := client.NetworkIDS(ctx)
	if err != nil {
		return nil, err
	}

	nonce, err := client.PendingNonceAtS(ctx, common.HexToAddress(txRequest.From))
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPriceS(ctx)
	if err != nil {
		return nil, err
	}

	info := &bsc.PreSignData{
		ChainID:  chainID.Int64(),
		Nonce:    nonce,
		GasPrice: gasPrice.Int64(),
	}

	switch txRequest.CoinType {
	case sphinxplugin.CoinType_CoinTypebinancecoin, sphinxplugin.CoinType_CoinTypetbinancecoin:
		info.GasLimit = 21_000
	case sphinxplugin.CoinType_CoinTypebinanceusd, sphinxplugin.CoinType_CoinTypetbinanceusd:
		info.ContractID = config.GetENV().Contract
		info.GasLimit = 300_000
	}

	out, err = json.Marshal(info)
	return out, err
}

// SendRawTransaction bsc
func SendRawTransaction(ctx context.Context, in []byte) (out []byte, err error) {
	signedData := &bsc.SignedData{}
	err = json.Unmarshal(in, signedData)
	if err != nil {
		return nil, err
	}
	client := bsc.Client()

	tx := new(types.Transaction)

	if err := rlp.Decode(bytes.NewReader(signedData.SignedTx), tx); err != nil {
		return nil, err
	}

	if err := client.SendTransactionS(ctx, tx); err != nil {
		return nil, err
	}

	broadcastedData := bsc.BroadcastedData{
		TxHash: tx.Hash().Hex(),
	}
	out, err = json.Marshal(broadcastedData)
	return out, err
}

// done(on chain) => true
func SyncTxState(ctx context.Context, in []byte) (out []byte, err error) {
	broadcastedData := &bsc.BroadcastedData{}
	err = json.Unmarshal(in, broadcastedData)
	if err != nil {
		return nil, err
	}
	return
	// client := bsc.Client()

	// _, isPending, err := client.TransactionByHashS(ctx, common.HexToHash(broadcastedData.TxHash))
	// if err != nil {
	// 	return false, err
	// }
	// if isPending {
	// 	return false, ErrWaitMessageOnChain
	// }

	// receipt, err := client.TransactionReceiptS(ctx, common.HexToHash(broadcastedData.TxHash))
	// if err != nil {
	// 	return false, err
	// }

	// logger.Sugar().Infof("transaction info: TxHash %v, GasUsed %v, Status %v.", receipt.TxHash, receipt.GasUsed, receipt.Status == 1)

	// if receipt.Status == 1 {
	// 	return true, nil
	// }

	// return false, ErrTransactionFail
}
