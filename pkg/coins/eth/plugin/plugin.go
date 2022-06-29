package plugin

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rlp"
)

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
	client := Client()

	if !common.IsHexAddress(addr) {
		return nil, ErrAddrNotValid
	}
	return client.BalanceAtS(ctx, common.HexToAddress(addr), nil)
}

func PreSign(ctx context.Context, coinType sphinxplugin.CoinType, from string) (*PreSignInfo, error) {
	client := Client()

	if !common.IsHexAddress(from) {
		return nil, ErrAddrNotValid
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
		ContractID: coins.USDTContract(chainID.Int64()),
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
		return false, ErrWaitMessageOnChain
	}

	receipt, err := client.TransactionReceiptS(ctx, common.HexToHash(txHash))
	if err != nil {
		return false, err
	}

	if receipt.Status == 1 {
		return true, nil
	}

	return false, ErrTransactionFail
}
