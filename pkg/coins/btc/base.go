package btc

import (
	"errors"
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
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

	ChainType       = sphinxplugin.ChainType_Bitcoin
	ChainNativeUnit = "BTC"
	ChainAtomicUnit = "Satoshi"
	ChainUnitExp    = 8
)

// BTCNetMap btc net map
var BTCNetMap = map[string]*chaincfg.Params{
	coins.CoinNetMain: &chaincfg.MainNetParams,
	coins.CoinNetTest: &chaincfg.RegressionNetParams,
}

// ErrWaitMessageOnChainMinConfirms ..
var ErrWaitMessageOnChainMinConfirms = errors.New("wait message on chain min confirms")

var (
	fundsTooLow    = `insufficient balance`
	listUnspendErr = `list unspent address fail`
	stopErrMsg     = []string{
		fundsTooLow,
		listUnspendErr,
		env.ErrEVNCoinNetValue.Error(),
		env.ErrAddressInvalid.Error(),
		env.ErrAmountInvalid.Error(),
	}
	bitcoinToken = &coins.TokenInfo{OfficialName: "Bitcoin", Decimal: 8, Unit: "BTC", Name: "bitcoin", OfficialContract: "bitcoin", TokenType: coins.Bitcoin}
)

func init() {
	// set chain info
	bitcoinToken.ChainType = ChainType
	bitcoinToken.ChainNativeUnit = ChainNativeUnit
	bitcoinToken.ChainAtomicUnit = ChainAtomicUnit
	bitcoinToken.ChainUnitExp = ChainUnitExp

	bitcoinToken.Waight = 100
	bitcoinToken.Net = coins.CoinNetMain
	bitcoinToken.Contract = bitcoinToken.OfficialContract
	bitcoinToken.CoinType = sphinxplugin.CoinType_CoinTypebitcoin
	register.RegisteTokenInfo(bitcoinToken)
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
