package bsc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
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
	client, err := Client()
	if err != nil {
		return nil, err
	}

	if !common.IsHexAddress(addr) {
		return nil, ErrAddrNotValid
	}
	return client.BalanceAtS(ctx, common.HexToAddress(addr), nil)
}

func PreSign(ctx context.Context, coinType sphinxplugin.CoinType, from string) (*PreSignInfo, error) {
	client, err := Client()
	if err != nil {
		return nil, err
	}

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

	gasLimit := int64(21_000)

	return &PreSignInfo{
		ChainID:  chainID.Int64(),
		Nonce:    nonce,
		GasPrice: gasPrice.Int64(),
		GasLimit: gasLimit,
	}, nil
}

// SendRawTransaction bsc
func SendRawTransaction(ctx context.Context, rawHexTx string) (string, error) {
	client, err := Client()
	if err != nil {
		return "", err
	}

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
	client, err := Client()
	if err != nil {
		return false, err
	}

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

	logger.Sugar().Infof("transaction info: TxHash %v, GasUsed %v, Status %v.", receipt.TxHash, receipt.GasUsed, receipt.Status == 1)

	if receipt.Status == 1 {
		return true, nil
	}

	return false, ErrTransactionFail
}