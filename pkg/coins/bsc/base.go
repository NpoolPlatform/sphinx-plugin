package bsc

import "errors"

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
