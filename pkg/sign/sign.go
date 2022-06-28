package sign

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

var (
	ErrCoinSignTypeAlreadyRegister = errors.New("coin sign type already register")
	ErrOpSignTypeAlreadyRegister   = errors.New("op sign type already register")

	ErrCoinSignTypeNotRegister = errors.New("coin sign type not register")
	ErrOpSignTypeNotRegister   = errors.New("op sign type not register")

	coinSignHandles = make(map[sphinxplugin.CoinType]map[sphinxplugin.TransactionType]Handlef)
)

type Handlef func(ctx context.Context, payload []byte) ([]byte, error)

func Register(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType, handle func(ctx context.Context, payload []byte) ([]byte, error)) {
	if opType != sphinxplugin.TransactionType_WalletNew && opType != sphinxplugin.TransactionType_Sign {
		panic(errors.New("??"))
	}
	coinPluginHandle, ok := coinSignHandles[coinType]
	if !ok {
		coinSignHandles[coinType] = make(map[sphinxplugin.TransactionType]Handlef)
	}
	if _, ok := coinPluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinSignHandles[coinType][opType] = handle
}

func GetCoinSign(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType) (func(ctx context.Context, payload []byte) ([]byte, error), bool) {
	if opType != sphinxplugin.TransactionType_WalletNew && opType != sphinxplugin.TransactionType_Sign {
		panic(errors.New("??"))
	}
	// TODO: check nested map exist
	if _, ok := coinSignHandles[coinType]; !ok {
		return nil, ok
	}
	if _, ok := coinSignHandles[coinType][opType]; !ok {
		return nil, ok
	}
	return coinSignHandles[coinType][opType], true
}
