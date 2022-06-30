package coins

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
)

var (
	ErrCoinTypeNotFound = errors.New("coin type not found")
	ErrOpTypeNotFound   = errors.New("op type not found")

	coinPluginHandles = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionState]Handlef)
)

// coin transaction handle
type Handlef func(ctx context.Context, payload []byte) ([]byte, error)

// register coin handle
// caution: not support dynamic register
func Register(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState, handle Handlef) {
	coinPluginHandle, ok := coinPluginHandles[coinType]
	if !ok {
		coinPluginHandles[coinType] = make(map[sphinxproxy.TransactionState]Handlef)
	}
	if _, ok := coinPluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinPluginHandles[coinType][opType] = handle
}

func GetCoinPlugin(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionState) (Handlef, error) {
	// TODO: check nested map exist
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
