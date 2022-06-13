package sign

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

var (
	ErrCoinSignTypeAlreadyRegister = errors.New("coin sign type already register")
	ErrOpSignTypeAlreadyRegister   = errors.New("op sign type already register")

	ErrCoinSignTypeNotRegister = errors.New("coin sign type not register")
	ErrOpSignTypeNotRegister   = errors.New("op sign type not register")

	coinSignHandles = make(map[sphinxplugin.CoinType]func(ctx context.Context, payload []byte) ([]byte, error))
)

func Register(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType, handle func(ctx context.Context, payload []byte) ([]byte, error)) {
	switch opType {
	case sphinxplugin.TransactionType_WalletNew:
		if _, ok := coinSignHandles[coinType]; ok {
			panic(ErrCoinSignTypeAlreadyRegister)
		}
		coinSignHandles[coinType] = handle
	case sphinxplugin.TransactionType_Sign:
		if _, ok := coinSignHandles[coinType]; ok {
			panic(ErrCoinSignTypeAlreadyRegister)
		}
		coinSignHandles[coinType] = handle
	default:
		panic(errors.New("!!>>??"))
	}
}

func GetCoinSign(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType) (func(ctx context.Context, payload []byte) ([]byte, error), bool) {
	switch opType {
	case sphinxplugin.TransactionType_WalletNew:
		handle, ok := coinSignHandles[coinType]
		return handle, ok
	case sphinxplugin.TransactionType_Sign:
		handle, ok := coinSignHandles[coinType]
		return handle, ok
	default:
		panic(errors.New("??"))
	}
}
