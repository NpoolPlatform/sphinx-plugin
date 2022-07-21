package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/log"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rlp"
)

// here register plugin func
func init() {
	// main
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypeethereum,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionType_Balance,
		walletBalance,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateWait,
		PreSign,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateBroadcast,
		SendRawTransaction,
	)
	coins.Register(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionState_TransactionStateSync,
		SyncTxState,
	)

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetethereum, eth.TxFailErr)
	if err != nil {
		panic(err)
	}

	coins.RegisterAbortErr(
		eth.ErrTransactionFail,
		eth.ErrAddrNotValid,
	)
}

func walletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return in, err
	}
	client := eth.Client()

	if !common.IsHexAddress(wbReq.Address) {
		return nil, eth.ErrAddrNotValid
	}
	bl, err := client.BalanceAtS(ctx, common.HexToAddress(wbReq.Address), nil)
	if err != nil {
		return in, err
	}

	balance, ok := big.NewFloat(0).SetString(bl.String())
	if !ok {
		return in, errors.New("convert balance string to float64 error")
	}

	balance.Quo(balance, big.NewFloat(math.Pow10(eth.ETHACCURACY)))
	f, exact := balance.Float64()
	if exact != big.Exact {
		log.Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	wbResp := &ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}
	out, err = json.Marshal(wbResp)

	return out, err
}

func PreSign(ctx context.Context, in []byte) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return in, err
	}
	client := eth.Client()

	if !common.IsHexAddress(baseInfo.From) {
		return in, eth.ErrAddrNotValid
	}

	chainID, err := client.NetworkIDS(ctx)
	if err != nil {
		return in, err
	}

	nonce, err := client.PendingNonceAtS(ctx, common.HexToAddress(baseInfo.From))
	if err != nil {
		return in, err
	}

	gasPrice, err := client.SuggestGasPriceS(ctx)
	if err != nil {
		return in, err
	}

	gasLimit := int64(0)
	contractID := ""
	switch baseInfo.CoinType {
	case sphinxplugin.CoinType_CoinTypeethereum, sphinxplugin.CoinType_CoinTypetethereum:
		gasLimit = 21_000
	case sphinxplugin.CoinType_CoinTypeusdterc20, sphinxplugin.CoinType_CoinTypetusdterc20:
		// client.EstimateGas(ctx, ethereum.CallMsg{})
		contractID = eth.USDTContract(chainID.Int64())
		gasLimit = 300_000
	}

	info := &eth.PreSignData{
		CoinType:   baseInfo.CoinType,
		ChainID:    chainID.Int64(),
		ContractID: contractID,
		Nonce:      nonce,
		GasPrice:   gasPrice.Int64(),
		GasLimit:   gasLimit,
		From:       baseInfo.From,
		To:         baseInfo.To,
		Value:      baseInfo.Value,
	}

	out, err = json.Marshal(info)
	if err != nil {
		return in, err
	}
	return out, err
}

// SendRawTransaction eth/usdt
func SendRawTransaction(ctx context.Context, in []byte) ([]byte, error) {
	signedData := &eth.SignedData{}
	err := json.Unmarshal(in, signedData)
	if err != nil {
		return in, err
	}
	logger.Sugar().Errorf("rlp: %v", signedData.SignedTx)
	client := eth.Client()

	tx := new(types.Transaction)

	if err := rlp.Decode(bytes.NewReader(signedData.SignedTx), tx); err != nil {
		return in, err
	}

	if err := client.SendTransactionS(ctx, tx); err != nil {
		return in, err
	}

	broadcastedData := ct.BroadcastInfo{
		TxID: tx.Hash().Hex(),
	}

	return json.Marshal(broadcastedData)
}

// SyncTxState done(on chain) => true
func SyncTxState(ctx context.Context, in []byte) (out []byte, err error) {
	broadcastedData := &ct.BroadcastInfo{}
	err = json.Unmarshal(in, broadcastedData)
	if err != nil {
		return in, err
	}
	client := eth.Client()

	_, isPending, err := client.TransactionByHashS(ctx, common.HexToHash(broadcastedData.TxID))
	if err != nil {
		return in, err
	}
	if isPending {
		return in, eth.ErrWaitMessageOnChain
	}

	receipt, err := client.TransactionReceiptS(ctx, common.HexToHash(broadcastedData.TxID))
	if err != nil {
		return in, err
	}
	log.Infof("transaction info: TxHash %v, GasUsed %v, Status %v.", receipt.TxHash, receipt.GasUsed, receipt.Status == 1)

	sResp := &ct.SyncResponse{ExitCode: 0}
	out, err = json.Marshal(sResp)
	if err != nil {
		return in, err
	}

	if receipt.Status == 1 {
		return out, nil
	}

	return in, eth.ErrTransactionFail
}
