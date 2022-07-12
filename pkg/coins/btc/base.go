package btc

import (
	"errors"
	"strings"

	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/btcsuite/btcd/chaincfg"
)

const (
	// BTCGas 0.00028BTC
	BTCGas = 0.00028
	// DefaultMinConfirms ..
	DefaultMinConfirms = 6
	// DefaultMaxConfirms ..
	DefaultMaxConfirms = 9999999
)

// BTCNetMap btc net map
var BTCNetMap = map[string]*chaincfg.Params{
	coins.CoinNetMain: &chaincfg.MainNetParams,
	coins.CoinNetTest: &chaincfg.RegressionNetParams,
}

// ErrWaitMessageOnChainMinConfirms ..
var ErrWaitMessageOnChainMinConfirms = errors.New("wait message on chain min confirms")

var (
	ErrFundsTooLow    = `insufficient balance`
	ErrListUnspendErr = `list unspent address fail`
	StopErrs          = []string{
		ErrFundsTooLow,
		ErrListUnspendErr,
		env.ErrEVNCoinNetValue.Error(),
		env.ErrAddressInvalid.Error(),
		env.ErrAmountInvalid.Error(),
	}
)

func TxFailErr(err error) bool {
	for _, v := range StopErrs {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
