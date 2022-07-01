package utils

import (
	"errors"
	"fmt"

	"github.com/NpoolPlatform/message/npool/sphinxplugin"
)

// ErrCoinTypeUnKnow ..
var ErrCoinTypeUnKnow = errors.New("coin type unknow")

const coinTypePrefix = "CoinType"

// ToCoinType ..
func ToCoinType(coinType string) (sphinxplugin.CoinType, error) {
	_coinType, ok := sphinxplugin.CoinType_value[fmt.Sprintf("%s%s", coinTypePrefix, coinType)]
	if !ok {
		return sphinxplugin.CoinType_CoinTypeUnKnow, ErrCoinTypeUnKnow
	}
	return sphinxplugin.CoinType(_coinType), nil
}
