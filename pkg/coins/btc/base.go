package btc

import (
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
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
