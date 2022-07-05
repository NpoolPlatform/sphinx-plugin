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

	coinPluginHandles        = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionState]Handlef)
	coinBalancePluginHandles = make(map[sphinxplugin.CoinType]map[sphinxproxy.TransactionType]Handlef)
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

func RegisterBalance(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionType, handle Handlef) {
	coinBalancePluginHandle, ok := coinBalancePluginHandles[coinType]
	if !ok {
		coinBalancePluginHandles[coinType] = make(map[sphinxproxy.TransactionType]Handlef)
	}
	if _, ok := coinBalancePluginHandle[opType]; ok {
		panic(fmt.Errorf("coin type: %v for transaction: %v already registered", coinType, opType))
	}
	coinBalancePluginHandles[coinType][opType] = handle
}

func GetCoinBalancePlugin(coinType sphinxplugin.CoinType, opType sphinxproxy.TransactionType) (Handlef, error) {
	// TODO: check nested map exist
	if _, ok := coinBalancePluginHandles[coinType]; !ok {
		return nil, ErrCoinTypeNotFound
	}
	if _, ok := coinBalancePluginHandles[coinType][opType]; !ok {
		return nil, ErrOpTypeNotFound
	}
	return coinBalancePluginHandles[coinType][opType], nil
}

// error ----------------------------
var (
	// ErrAbortErrorAlreadyRegister ..
	ErrAbortErrorAlreadyRegister = errors.New("abort error already register")

	// TODO: think how to check not value error
	abortErrs = make(map[error]struct{})
)

// RegisterAbortErr ..
func RegisterAbortErr(errs ...error) error {
	for _, err := range errs {
		if _, ok := abortErrs[err]; ok {
			// return ErrAbortErrorAlreadyRegister
			continue
		}
		abortErrs[err] = struct{}{}
	}

	return nil
}

func nextStop(err error) bool {
	if err == nil {
		return false
	}

	_, ok := abortErrs[err]
	return ok
}

// NextStop ..
func NextStop(err error) bool {
	return nextStop(err)
}
