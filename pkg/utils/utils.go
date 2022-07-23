package utils

import (
	"errors"
	"fmt"
	"math"

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

func MinInt(args ...int) int {
	minNum := math.MaxInt
	for _, v := range args {
		if v < minNum {
			minNum = v
		}
	}
	return minNum
}
