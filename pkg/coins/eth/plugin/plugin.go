package plugin

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/eth"
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
		WalletBalance,
	)
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypeethereum,
	// 	sphinxproxy.TransactionState_TransactionStateWait,
	// 	PreSign,
	// )
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypeethereum,
	// 	sphinxproxy.TransactionState_TransactionStateBroadcast,
	// 	SendRawTransaction,
	// )
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypeethereum,
	// 	sphinxproxy.TransactionState_TransactionStateSync,
	// 	SyncTxState,
	// )

	// test
	coins.RegisterBalance(
		sphinxplugin.CoinType_CoinTypetethereum,
		sphinxproxy.TransactionType_Balance,
		WalletBalance,
	)
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypetethereum,
	// 	sphinxproxy.TransactionState_TransactionStateWait,
	// 	PreSign,
	// )
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypetethereum,
	// 	sphinxproxy.TransactionState_TransactionStateBroadcast,
	// 	SendRawTransaction,
	// )
	// coins.Register(
	// 	sphinxplugin.CoinType_CoinTypetethereum,
	// 	sphinxproxy.TransactionState_TransactionStateSync,
	// 	SyncTxState,
	// )

	err := coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypeethereum, IsErrStop)
	if err != nil {
		panic(err)
	}

	err = coins.RegisterAbortFuncErr(sphinxplugin.CoinType_CoinTypetethereum, IsErrStop)
	if err != nil {
		panic(err)
	}

	coins.RegisterAbortErr(eth.ErrTransactionFail)
}

type PreSignInfo struct {
	ChainID    int64
	ContractID string
	Nonce      uint64
	GasPrice   int64
	GasLimit   int64
}

func WalletBalance(ctx context.Context, in []byte) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return in, err
	}
	client := Client()

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
		logger.Sugar().Warnf("wallet balance transfer warning balance from->to %v-%v", balance.String(), f)
	}

	wbResp := &ct.WalletBalanceResponse{
		Balance:    f,
		BalanceStr: balance.String(),
	}
	out, err = json.Marshal(wbResp)

	return out, err
}

func PreSign(ctx context.Context, coinType sphinxplugin.CoinType, from string) (*PreSignInfo, error) {
	client := Client()

	if !common.IsHexAddress(from) {
		return nil, eth.ErrAddrNotValid
	}

	chainID, err := client.NetworkIDS(ctx)
	if err != nil {
		return nil, err
	}

	nonce, err := client.PendingNonceAtS(ctx, common.HexToAddress(from))
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPriceS(ctx)
	if err != nil {
		return nil, err
	}

	gasLimit := int64(0)

	switch coinType {
	case sphinxplugin.CoinType_CoinTypeethereum, sphinxplugin.CoinType_CoinTypetethereum:
		gasLimit = 21_000
	case sphinxplugin.CoinType_CoinTypeusdterc20, sphinxplugin.CoinType_CoinTypetusdterc20:
		// client.EstimateGas(ctx, ethereum.CallMsg{})
		gasLimit = 300_000
	}

	return &PreSignInfo{
		ChainID:    chainID.Int64(),
		ContractID: eth.USDTContract(chainID.Int64()),
		Nonce:      nonce,
		GasPrice:   gasPrice.Int64(),
		GasLimit:   gasLimit,
	}, nil
}

// SendRawTransaction eth/usdt
func SendRawTransaction(ctx context.Context, rawHexTx string) (string, error) {
	client := Client()

	tx := new(types.Transaction)

	rawByteTx, err := hex.DecodeString(rawHexTx)
	if err != nil {
		return "", err
	}

	if err := rlp.Decode(bytes.NewReader(rawByteTx), tx); err != nil {
		return "", err
	}

	if err := client.SendTransactionS(ctx, tx); err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

// done(on chain) => true
func SyncTxState(ctx context.Context, txHash string) (bool, error) {
	client := Client()

	_, isPending, err := client.TransactionByHashS(ctx, common.HexToHash(txHash))
	if err != nil {
		return false, err
	}
	if isPending {
		return false, eth.ErrWaitMessageOnChain
	}

	receipt, err := client.TransactionReceiptS(ctx, common.HexToHash(txHash))
	if err != nil {
		return false, err
	}

	if receipt.Status == 1 {
		return true, nil
	}

	return false, eth.ErrTransactionFail
}

func IsErrStop(err error) bool {
	if err.Error() == "" {
		return false
	}
	matchedErrs := []string{
		`intrinsic gas too low`,                      // gas low
		`insufficient funds for gas * price + value`, // funds low
	}

	for _, v := range matchedErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}

	return false
}
