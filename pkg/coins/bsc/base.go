package bsc

import (
	"errors"
	"strings"
)

const (
	BNBACCURACY   = 18
	BEP20ACCURACY = 18
)

var (
	// ErrWaitMessageOnChain ..
	ErrWaitMessageOnChain = errors.New("wait message on chain")
	// ErrAddrNotValid ..
	ErrAddrNotValid = errors.New("invalid bsc address")
	// ErrTransactionFail ..
	ErrTransactionFail = errors.New("bsc transaction fail")
)

var (
	ErrGasTooLow   = `intrinsic gas too low`
	ErrFundsTooLow = `insufficient funds for gas * price + value`
	ErrNonceToLow  = `nonce too low`
	StopErrs       = []string{ErrGasTooLow, ErrFundsTooLow, ErrNonceToLow}
)

func TxFailErr(err error) bool {
	for _, v := range StopErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
