package plugin

import (
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/filecoin-project/go-address"
)

const (
	CoinNetMain = "main"
	CoinNetTest = "test"

	// DefaultMinConfirms ..
	DefaultMinConfirms = 6
)

var (
	// CoinNet will filled value in app run
	CoinNet string

	CoinUnit = map[string]map[sphinxplugin.CoinType]string{
		// main
		CoinNetMain: {
			sphinxplugin.CoinType_CoinTypefilecoin: "FIL",
			sphinxplugin.CoinType_CoinTypebitcoin:  "BTC",
			sphinxplugin.CoinType_CoinTypeethereum: "ETH",
		},

		// test
		CoinNetTest: {
			sphinxplugin.CoinType_CoinTypefilecoin: "tFIL",
			sphinxplugin.CoinType_CoinTypebitcoin:  "tBTC",
			sphinxplugin.CoinType_CoinTypeethereum: "tETH",
		},
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
)

func CheckSupportNet(netEnv string) bool {
	return (netEnv == CoinNetMain ||
		netEnv == CoinNetTest)
}

// TODO match case elegant deal
func CoinStr2CoinType(coinStr string) sphinxplugin.CoinType {
	switch strings.ToLower(coinStr) {
	case "fil":
		return sphinxplugin.CoinType_CoinTypefilecoin
	case "btc":
		return sphinxplugin.CoinType_CoinTypebitcoin
	case "eth":
		return sphinxplugin.CoinType_CoinTypeethereum
	case "spacemesh":
		return sphinxplugin.CoinType_CoinTypespacemesh
	default:
	}

	return sphinxplugin.CoinType_CoinTypeUnKnow
}
