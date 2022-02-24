package plugin

import (
	"strings"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

const (
	// DefaultMinConfirms ..
	DefaultMinConfirms = 6
)

var CoinUnit = map[sphinxplugin.CoinType]string{
	sphinxplugin.CoinType_CoinTypefilecoin: "FIL",
	sphinxplugin.CoinType_CoinTypebitcoin:  "BTC",
	sphinxplugin.CoinType_CoinTypeethereum: "ETH",
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
