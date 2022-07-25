package coins

import (
	"context"
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/message/npool/sphinxproxy"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
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

	// ErrAbortErrorFuncAlreadyRegister ..
	ErrAbortErrorFuncAlreadyRegister = errors.New("abort error func already register")

	// TODO: think how to check not value error
	abortErrs = map[error]struct{}{
		env.ErrEVNCoinNet:      {},
		env.ErrEVNCoinNetValue: {},
		env.ErrAddressInvalid:  {},
		env.ErrSignTypeInvalid: {},
		env.ErrCIDInvalid:      {},
		env.ErrContractInvalid: {},
		env.ErrTransactionFail: {},
	}

	abortFuncErrs = make(map[sphinxplugin.CoinType]func(error) bool)
)

// RegisterAbortErr ..
func RegisterAbortErr(errs ...error) {
	for _, err := range errs {
		if _, ok := abortErrs[err]; ok {
			panic(ErrAbortErrorAlreadyRegister)
		}
		abortErrs[err] = struct{}{}
	}
}

// RegisterAbortFuncErr ..
func RegisterAbortFuncErr(coinType sphinxplugin.CoinType, f func(error) bool) error {
	if _, ok := abortFuncErrs[coinType]; ok {
		return ErrAbortErrorFuncAlreadyRegister
	}

	abortFuncErrs[coinType] = f
	return nil
}

func nextStop(err error) bool {
	if err == nil {
		return false
	}

	_, ok := abortErrs[err]
	return ok
}

// Abort ..
func Abort(coinType sphinxplugin.CoinType, err error) bool {
	if err == nil {
		return false
	}

	if nextStop(err) {
		return true
	}

	mf, ok := abortFuncErrs[coinType]
	if ok {
		return mf(err)
	}

	return false
}
