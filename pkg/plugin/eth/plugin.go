package eth

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rlp"
)

func WalletBalance(ctx context.Context, addr string) (*big.Int, error) {
	client, err := client()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.BalanceAt(ctx, common.HexToAddress(addr), nil)
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
