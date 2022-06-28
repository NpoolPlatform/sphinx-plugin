package coins

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

var (
	ErrCoinTypeNotFound = errors.New("coin type not found")
	ErrOpTypeNotFound   = errors.New("op type not found")

	coinPluginHandles = map[sphinxplugin.CoinType]map[sphinxplugin.TransactionType]Handlef{}
)

// coin transaction handle
type Handlef func(ctx context.Context, payload []byte) ([]byte, error)

// register coin handle
// caution: not support dynamic register
func Register(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType, handle Handlef) {
	if _, ok := coinPluginHandles[coinType][opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registed", coinType, opType))
	}
	coinPluginHandles[coinType][opType] = handle
}

func GetCoinPlugin(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType) (Handlef, error) {
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
