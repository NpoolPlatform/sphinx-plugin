package depinc

import (
	"errors"
	"strings"

	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/env"
	"github.com/btcsuite/btcd/chaincfg"
)

const (
	// DEPCGas 0.00028DEPC
	// the minimum value is
	DEPCGas = 0.00000280
	// DefaultMinConfirms ..
	DefaultMinConfirms = 6
	// DefaultMaxConfirms ..
	DefaultMaxConfirms = 9999999

	ChainType           = sphinxplugin.ChainType_Depinc
	ChainNativeUnit     = "DePC"
	ChainOfficialName   = "Depinc"
	ChainAtomicUnit     = "Satoshi (sat)"
	ChainUnitExp        = 8
	ChainNativeCoinName = "depinc"
	ChainID             = "N/A"
)

// DEPCNetMap depinc net map
var DEPCNetMap = map[string]*chaincfg.Params{
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
	depincToken = &coins.TokenInfo{OfficialName: ChainOfficialName, Decimal: ChainUnitExp, Unit: ChainNativeUnit, Name: ChainNativeCoinName, OfficialContract: ChainNativeCoinName, TokenType: coins.Depinc}
)

func init() {
	// set chain info
	depincToken.ChainType = ChainType
	depincToken.ChainNativeUnit = ChainNativeUnit
	depincToken.ChainAtomicUnit = ChainAtomicUnit
	depincToken.ChainUnitExp = ChainUnitExp
	depincToken.GasType = v1.GasType_GasUnsupported
	depincToken.ChainID = ChainID
	depincToken.ChainNickname = ChainType.String()
	depincToken.ChainNativeCoinName = ChainNativeCoinName

	depincToken.Waight = 100
	depincToken.Net = coins.CoinNetMain
	depincToken.Contract = depincToken.OfficialContract
	depincToken.CoinType = sphinxplugin.CoinType_CoinTypedepinc
	register.RegisteTokenInfo(depincToken)
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
