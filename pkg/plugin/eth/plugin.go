package eth

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rlp"
)

var ErrWaitMessageOnChain = errors.New("wait message on chain")

type PreSignInfo struct {
	ChainID  int64
	Nonce    uint64
	GasPrice int64
	GasLimit int64
}

func WalletBalance(ctx context.Context, addr string) (*big.Int, error) {
	client, err := client()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.BalanceAt(ctx, common.HexToAddress(addr), nil)
}

func PreSign(ctx context.Context, coinType sphinxplugin.CoinType, from string) (*PreSignInfo, error) {
	client, err := client()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(from))
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	gasLimit := int64(0)

	switch coinType {
	case sphinxplugin.CoinType_CoinTypeethereum, sphinxplugin.CoinType_CoinTypetethereum:
		gasLimit = 300000
	case sphinxplugin.CoinType_CoinTypeusdterc20, sphinxplugin.CoinType_CoinTypetusdterc20:
		// client.EstimateGas(ctx, ethereum.CallMsg{})
		gasLimit = 300000
	}

	return &PreSignInfo{
		ChainID:  chainID.Int64(),
		Nonce:    nonce,
		GasPrice: gasPrice.Int64(),
		GasLimit: gasLimit,
	}, nil
}

// SendRawTransaction eth/usdt
func SendRawTransaction(ctx context.Context, rawHexTx string) (string, error) {
	client, err := client()
	if err != nil {
		return "", err
	}
	defer client.Close()

	tx := new(types.Transaction)

	if err := rlp.Decode(strings.NewReader(rawHexTx), tx); err != nil {
		return "", err
	}

	if err := client.SendTransaction(ctx, tx); err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

// done(on chain) => true
func SyncTxState(ctx context.Context, txHash string) (bool, error) {
	client, err := client()
	if err != nil {
		return false, err
	}
	defer client.Close()

	_, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return false, err
	}

	return !isPending, nil
}
