package coins

import (
	"context"
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

var (
	ErrCoinTypeAlreadyRegister = errors.New("coin type already register")
	ErrOpTypeAlreadyRegister   = errors.New("op type already register")

	coinPluginHandles = make(map[sphinxplugin.CoinType]map[sphinxplugin.TransactionType]func(ctx context.Context, payload []byte) ([]byte, error))
)

// register coin handle
func Register(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType, handle func(ctx context.Context, payload []byte) ([]byte, error)) {
	if coinHandle, ok := coinPluginHandles[coinType]; ok {
		panic(ErrCoinTypeAlreadyRegister)
	} else if _, ok := coinHandle[opType]; ok {
		panic(ErrOpTypeAlreadyRegister)
	}
	coinPluginHandles[coinType][opType] = handle
}

func GetCoinPlugin(coinType sphinxplugin.CoinType, opType sphinxplugin.TransactionType) func(ctx context.Context, payload []byte) ([]byte, error) {
	// TODO: check nested map exist
	return coinPluginHandles[coinType][opType]
}

type IPlugin interface {
	// NewAccount(ctx context.Context, req []byte) ([]byte, error)
	// Sign(ctx context.Context, req []byte) ([]byte, error)

	WalletBalance(ctx context.Context, req []byte) ([]byte, error)
	PreSign(ctx context.Context, req []byte) ([]byte, error)
	Broadcast(ctx context.Context, req []byte) ([]byte, error)
	SyncTx(ctx context.Context, req []byte) error
}
