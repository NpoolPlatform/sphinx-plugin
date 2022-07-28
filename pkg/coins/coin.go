package coins

import (
	"strings"
	"time"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

const (
	CoinNetMain = "main"
	CoinNetTest = "test"
)

var (
	// not export
	netCoinMap = map[string]map[string]sphinxplugin.CoinType{
		CoinNetMain: {
			"filecoin":    sphinxplugin.CoinType_CoinTypefilecoin,
			"bitcoin":     sphinxplugin.CoinType_CoinTypebitcoin,
			"ethereum":    sphinxplugin.CoinType_CoinTypeethereum,
			"usdterc20":   sphinxplugin.CoinType_CoinTypeusdterc20,
			"spacemesh":   sphinxplugin.CoinType_CoinTypespacemesh,
			"solana":      sphinxplugin.CoinType_CoinTypesolana,
			"usdttrc20":   sphinxplugin.CoinType_CoinTypeusdttrc20,
			"tron":        sphinxplugin.CoinType_CoinTypetron,
			"binancecoin": sphinxplugin.CoinType_CoinTypebinancecoin,
			"binanceusd":  sphinxplugin.CoinType_CoinTypebinanceusd,
			"usdcerc20":   sphinxplugin.CoinType_CoinTypeusdcerc20,
		},
		CoinNetTest: {
			"filecoin":    sphinxplugin.CoinType_CoinTypetfilecoin,
			"bitcoin":     sphinxplugin.CoinType_CoinTypetbitcoin,
			"ethereum":    sphinxplugin.CoinType_CoinTypetethereum,
			"usdterc20":   sphinxplugin.CoinType_CoinTypetusdterc20,
			"spacemesh":   sphinxplugin.CoinType_CoinTypetspacemesh,
			"solana":      sphinxplugin.CoinType_CoinTypetsolana,
			"usdttrc20":   sphinxplugin.CoinType_CoinTypetusdttrc20,
			"tron":        sphinxplugin.CoinType_CoinTypettron,
			"binancecoin": sphinxplugin.CoinType_CoinTypetbinancecoin,
			"binanceusd":  sphinxplugin.CoinType_CoinTypetbinanceusd,
			"usdcerc20":   sphinxplugin.CoinType_CoinTypetusdcerc20,
		},
	}

	// not export
	coinNetMap = map[sphinxplugin.CoinType]string{
		// main
		sphinxplugin.CoinType_CoinTypefilecoin:    CoinNetMain,
		sphinxplugin.CoinType_CoinTypebitcoin:     CoinNetMain,
		sphinxplugin.CoinType_CoinTypeethereum:    CoinNetMain,
		sphinxplugin.CoinType_CoinTypeusdterc20:   CoinNetMain,
		sphinxplugin.CoinType_CoinTypespacemesh:   CoinNetMain,
		sphinxplugin.CoinType_CoinTypesolana:      CoinNetMain,
		sphinxplugin.CoinType_CoinTypeusdttrc20:   CoinNetMain,
		sphinxplugin.CoinType_CoinTypetron:        CoinNetMain,
		sphinxplugin.CoinType_CoinTypebinancecoin: CoinNetMain,
		sphinxplugin.CoinType_CoinTypebinanceusd:  CoinNetMain,
		sphinxplugin.CoinType_CoinTypeusdcerc20:   CoinNetMain,

		// test
		sphinxplugin.CoinType_CoinTypetfilecoin:    CoinNetTest,
		sphinxplugin.CoinType_CoinTypetbitcoin:     CoinNetTest,
		sphinxplugin.CoinType_CoinTypetethereum:    CoinNetTest,
		sphinxplugin.CoinType_CoinTypetusdterc20:   CoinNetTest,
		sphinxplugin.CoinType_CoinTypetspacemesh:   CoinNetTest,
		sphinxplugin.CoinType_CoinTypetsolana:      CoinNetTest,
		sphinxplugin.CoinType_CoinTypetusdttrc20:   CoinNetTest,
		sphinxplugin.CoinType_CoinTypettron:        CoinNetTest,
		sphinxplugin.CoinType_CoinTypetbinancecoin: CoinNetTest,
		sphinxplugin.CoinType_CoinTypetbinanceusd:  CoinNetTest,
		sphinxplugin.CoinType_CoinTypetusdcerc20:   CoinNetTest,
	}

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

		sphinxplugin.CoinType_CoinTypebinancecoin:  "BNB",
		sphinxplugin.CoinType_CoinTypetbinancecoin: "BNB",

		sphinxplugin.CoinType_CoinTypebinanceusd:  "BUSD",
		sphinxplugin.CoinType_CoinTypetbinanceusd: "BUSD",

		sphinxplugin.CoinType_CoinTypeusdcerc20:  "USDC",
		sphinxplugin.CoinType_CoinTypetusdcerc20: "USDC",
	}

	// default sync time for waitting transaction on chain
	SyncTime = map[sphinxplugin.CoinType]time.Duration{
		sphinxplugin.CoinType_CoinTypefilecoin:  time.Second * 20,
		sphinxplugin.CoinType_CoinTypetfilecoin: time.Second * 20,

		sphinxplugin.CoinType_CoinTypebitcoin:  time.Minute * 7,
		sphinxplugin.CoinType_CoinTypetbitcoin: time.Minute * 7,

		sphinxplugin.CoinType_CoinTypeethereum:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetethereum: time.Second * 12,

		sphinxplugin.CoinType_CoinTypeusdterc20:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetusdterc20: time.Second * 12,

		sphinxplugin.CoinType_CoinTypeusdcerc20:  time.Second * 12,
		sphinxplugin.CoinType_CoinTypetusdcerc20: time.Second * 12,

		sphinxplugin.CoinType_CoinTypesolana:  time.Second * 1,
		sphinxplugin.CoinType_CoinTypetsolana: time.Second * 1,

		sphinxplugin.CoinType_CoinTypeusdttrc20:  time.Second * 2,
		sphinxplugin.CoinType_CoinTypetusdttrc20: time.Second * 2,

		sphinxplugin.CoinType_CoinTypetron:  time.Second * 2,
		sphinxplugin.CoinType_CoinTypettron: time.Second * 2,

		sphinxplugin.CoinType_CoinTypebinancecoin:  time.Second * 4,
		sphinxplugin.CoinType_CoinTypetbinancecoin: time.Second * 4,

		sphinxplugin.CoinType_CoinTypebinanceusd:  time.Second * 4,
		sphinxplugin.CoinType_CoinTypetbinanceusd: time.Second * 4,
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
