package eth

import (
	"errors"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
)

const (
	ETHACCURACY   = 18
	ERC20ACCURACY = 6
)

// USDTContract ...
var USDTContract = func(chainet int64) string {
	switch chainet {
	case 1:
		return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	case 1337:
		contract, ok := env.LookupEnv(env.ENVCONTRACT)
		if !ok {
			panic(env.ErrENVContractNotFound)
		}
		return contract
	}
	return ""
}

var (
	// ErrWaitMessageOnChain ..
	ErrWaitMessageOnChain = errors.New("wait message on chain")
	// ErrAddrNotValid ..
	ErrAddrNotValid = errors.New("invalid eth address")
	// ErrTransactionFail ..
	ErrTransactionFail = errors.New("eth transaction fail")
)
