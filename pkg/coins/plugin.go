package coins

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

var (
	ErrCoinTypeAlreadyRegister = errors.New("coin type already register")
	ErrOpTypeAlreadyRegister   = errors.New("op type already register")
	ErrCoinTypeNotFound        = errors.New("coin type not found")
	ErrOpTypeNotFound          = errors.New("op type not found")
	coinPluginHandles          = make(map[sphinxplugin.CoinType]map[sphinxplugin.TransactionType]func(ctx context.Context, payload []byte) ([]byte, error))
)

// register coin handle
func Register(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType, handle func(ctx context.Context, payload []byte) ([]byte, error)) {
	if _, ok := coinPluginHandles[coinType]; !ok {
		coinPluginHandles[coinType] = make(map[sphinxplugin.TransactionType]func(ctx context.Context, payload []byte) ([]byte, error))
	}
	if _, ok := coinPluginHandles[coinType][opType]; ok {
		panic(ErrOpTypeAlreadyRegister)
	}
	coinPluginHandles[coinType][opType] = handle
}

func GetCoinPlugin(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType) (func(ctx context.Context, payload []byte) ([]byte, error), error) {
	// TODO: check nested map exist
	// DO
	if _, ok := coinPluginHandles[coinType]; !ok {
		return nil, ErrCoinTypeNotFound
	}
	if _, ok := coinPluginHandles[coinType][opType]; !ok {
		return nil, ErrOpTypeNotFound
	}
	return coinPluginHandles[coinType][opType], nil
}

type IPlugin interface {
	// NewAccount(ctx context.Context, req []byte) ([]byte, error)
	// Sign(ctx context.Context, req []byte) ([]byte, error)

	WalletBalance(ctx context.Context, req []byte) ([]byte, error)
	PreSign(ctx context.Context, req []byte) ([]byte, error)
	Broadcast(ctx context.Context, req []byte) ([]byte, error)
	SyncTx(ctx context.Context, req []byte) error
}
