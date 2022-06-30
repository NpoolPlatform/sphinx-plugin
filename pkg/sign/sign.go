package sign

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
)

var (
	ErrCoinSignTypeAlreadyRegister = errors.New("coin sign type already register")
	ErrOpSignTypeAlreadyRegister   = errors.New("op sign type already register")

	ErrCoinSignTypeNotRegister = errors.New("coin sign type not register")
	ErrOpSignTypeNotRegister   = errors.New("op sign type not register")

	coinSignHandles = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionState]Handlef)
)

type Handlef func(ctx context.Context, payload []byte) ([]byte, error)

func Register(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState, handle func(ctx context.Context, payload []byte) ([]byte, error)) {
	if opType != sphinxproxy.TransactionState_TransactionStateSign {
		panic(errors.New("??"))
	}
	coinPluginHandle, ok := coinSignHandles[coinType]
	if !ok {
		coinSignHandles[coinType] = make(map[sphinxproxy.TransactionState]Handlef)
	}
	if _, ok := coinPluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinSignHandles[coinType][opType] = handle
}

func GetCoinSign(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState) (func(ctx context.Context, payload []byte) ([]byte, error), bool) {
	if opType != sphinxproxy.TransactionState_TransactionStateSign {
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
