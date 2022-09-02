package coins

import (
	"errors"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

var (
	ErrCoinTypeNotFound = errors.New("coin type not found")
	ErrOpTypeNotFound   = errors.New("op type not found")
)

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
