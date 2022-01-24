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
	sphinxplugin.CoinType_CoinTypeFIL: "FIL",
	sphinxplugin.CoinType_CoinTypeBTC: "BTC",
}

func CoinStr2CoinType(coinStr string) sphinxplugin.CoinType {
	switch strings.ToLower(coinStr) {
	case "fil":
		return sphinxplugin.CoinType_CoinTypeFIL
	case "btc":
		return sphinxplugin.CoinType_CoinTypeBTC
	default:
	}

	return sphinxplugin.CoinType_CoinTypeUnKnow
}
