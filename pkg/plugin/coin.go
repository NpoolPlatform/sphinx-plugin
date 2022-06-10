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
			"filecoin":  sphinxplugin.CoinType_CoinTypefilecoin,
			"bitcoin":   sphinxplugin.CoinType_CoinTypebitcoin,
			"ethereum":  sphinxplugin.CoinType_CoinTypeethereum,
			"usdterc20": sphinxplugin.CoinType_CoinTypeusdterc20,
			"spacemesh": sphinxplugin.CoinType_CoinTypespacemesh,
			"solana":    sphinxplugin.CoinType_CoinTypesolana,
			"usdttrc20": sphinxplugin.CoinType_CoinTypeusdttrc20,
			"tron":      sphinxplugin.CoinType_CoinTypetron,
		},
		CoinNetTest: {
			"filecoin":  sphinxplugin.CoinType_CoinTypetfilecoin,
			"bitcoin":   sphinxplugin.CoinType_CoinTypetbitcoin,
			"ethereum":  sphinxplugin.CoinType_CoinTypetethereum,
			"usdterc20": sphinxplugin.CoinType_CoinTypetusdterc20,
			"spacemesh": sphinxplugin.CoinType_CoinTypetspacemesh,
			"solana":    sphinxplugin.CoinType_CoinTypetsolana,
			"usdttrc20": sphinxplugin.CoinType_CoinTypetusdttrc20,
			"tron":      sphinxplugin.CoinType_CoinTypettron,
		},
	}

	// not export
	coinNetMap = map[sphinxplugin.CoinType]string{
		// main
		sphinxplugin.CoinType_CoinTypefilecoin:  CoinNetMain,
		sphinxplugin.CoinType_CoinTypebitcoin:   CoinNetMain,
		sphinxplugin.CoinType_CoinTypeethereum:  CoinNetMain,
		sphinxplugin.CoinType_CoinTypeusdterc20: CoinNetMain,
		sphinxplugin.CoinType_CoinTypespacemesh: CoinNetMain,
		sphinxplugin.CoinType_CoinTypesolana:    CoinNetMain,
		sphinxplugin.CoinType_CoinTypeusdttrc20: CoinNetMain,
		sphinxplugin.CoinType_CoinTypetron:      CoinNetMain,

		// test
		sphinxplugin.CoinType_CoinTypetfilecoin:  CoinNetTest,
		sphinxplugin.CoinType_CoinTypetbitcoin:   CoinNetTest,
		sphinxplugin.CoinType_CoinTypetethereum:  CoinNetTest,
		sphinxplugin.CoinType_CoinTypetusdterc20: CoinNetTest,
		sphinxplugin.CoinType_CoinTypetspacemesh: CoinNetTest,
		sphinxplugin.CoinType_CoinTypetsolana:    CoinNetTest,
		sphinxplugin.CoinType_CoinTypetusdttrc20: CoinNetTest,
		sphinxplugin.CoinType_CoinTypettron:      CoinNetTest,
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

		sphinxplugin.CoinType_CoinTypeusdterc20:  "USDT",
		sphinxplugin.CoinType_CoinTypetusdterc20: "USDT",

		sphinxplugin.CoinType_CoinTypesolana:  "SOL",
		sphinxplugin.CoinType_CoinTypetsolana: "SOL",

		sphinxplugin.CoinType_CoinTypeusdttrc20:  "USDT",
		sphinxplugin.CoinType_CoinTypetusdttrc20: "USDT",

		sphinxplugin.CoinType_CoinTypetron:  "TRX",
		sphinxplugin.CoinType_CoinTypettron: "TRX",
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

	// usdt contract
	USDTContract = func(chainet int64) string {
		switch chainet {
		case 1:
			return "0xdAC17F958D2ee523a2206206994597C13D831ec7"
		case 1337:
			return config.GetENV().Contract
		}
		return ""
	}
)

// CoinInfo report coin info
type CoinInfo struct {
	ENV      string // main or test
	Unit     string
	IP       string // wan ip
	Location string
}

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
