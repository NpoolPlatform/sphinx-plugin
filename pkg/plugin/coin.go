package plugin

import (
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/NpoolPlatform/sphinx-plugin/pkg/config"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/filecoin-project/go-address"
)

const (
	CoinNetMain = "main"
	CoinNetTest = "test"

	// DefaultMinConfirms ..
	DefaultMinConfirms = 6
	// DefaultMaxConfirms ..
	DefaultMaxConfirms = 9999999
)

var (
	// not export
	netCoinMap = map[string]map[string]sphinxplugin.CoinType{
		CoinNetMain: {
			"fil":       sphinxplugin.CoinType_CoinTypefilecoin,
			"btc":       sphinxplugin.CoinType_CoinTypebitcoin,
			"eth":       sphinxplugin.CoinType_CoinTypeethereum,
			"spacemesh": sphinxplugin.CoinType_CoinTypespacemesh,
		},
		CoinNetTest: {
			"fil":       sphinxplugin.CoinType_CoinTypetfilecoin,
			"btc":       sphinxplugin.CoinType_CoinTypetbitcoin,
			"eth":       sphinxplugin.CoinType_CoinTypetethereum,
			"spacemesh": sphinxplugin.CoinType_CoinTypetspacemesh,
		},
	}

	// not export
	coinNetMap = map[sphinxplugin.CoinType]string{
		// main
		sphinxplugin.CoinType_CoinTypefilecoin:  CoinNetMain,
		sphinxplugin.CoinType_CoinTypebitcoin:   CoinNetMain,
		sphinxplugin.CoinType_CoinTypeethereum:  CoinNetMain,
		sphinxplugin.CoinType_CoinTypespacemesh: CoinNetMain,

		// test
		sphinxplugin.CoinType_CoinTypetfilecoin:  CoinNetTest,
		sphinxplugin.CoinType_CoinTypetbitcoin:   CoinNetTest,
		sphinxplugin.CoinType_CoinTypetethereum:  CoinNetTest,
		sphinxplugin.CoinType_CoinTypetspacemesh: CoinNetTest,
	}

	// CoinNet will filled value in app run
	CoinNet string

	CoinUnit = map[sphinxplugin.CoinType]string{
		sphinxplugin.CoinType_CoinTypefilecoin:  "FIL",
		sphinxplugin.CoinType_CoinTypetfilecoin: "FIL",

		sphinxplugin.CoinType_CoinTypebitcoin:  "BTC",
		sphinxplugin.CoinType_CoinTypetbitcoin: "BTC",

		sphinxplugin.CoinType_CoinTypeethereum:  "ETH",
		sphinxplugin.CoinType_CoinTypetethereum: "ETH",
	}

	// BTCNetMap btc net map
	BTCNetMap = map[string]*chaincfg.Params{
		CoinNetMain: &chaincfg.MainNetParams,
		CoinNetTest: &chaincfg.RegressionNetParams,
	}

	// FILNetMap fil net map
	FILNetMap = map[string]address.Network{
		CoinNetMain: address.Mainnet,
		CoinNetTest: address.Testnet,
	}

	// usdt contract id
	USDTContractID = func(chainet int64) string {
		switch chainet {
		case 1:
			return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
		case 1337:
			return config.GetString(config.KeyContractID)
		}
		return ""
	}
)

// CoinType2Net ..
func CoinType2Net(ct sphinxplugin.CoinType) string {
	return coinNetMap[ct]
}

// CheckSupportNet ..
func CheckSupportNet(netEnv string) bool {
	return (netEnv == CoinNetMain ||
		netEnv == CoinNetTest)
}

// TODO match case elegant deal
func CoinStr2CoinType(netEnv, coinStr string) sphinxplugin.CoinType {
	_netEnv := strings.ToLower(netEnv)
	_coinStr := strings.ToLower(coinStr)
	return netCoinMap[_netEnv][_coinStr]
}
